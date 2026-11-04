// Package xid generates globally unique, time-sortable 20-character IDs in the
// same format as github.com/rs/xid: 12 bytes (4-byte big-endian Unix seconds,
// 3-byte machine ID, 2-byte PID, 3-byte counter) encoded as lowercase
// base32hex without padding.
package xid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"os"
	"sync/atomic"
	"time"
)

// ID is a 12-byte unique identifier.
type ID [12]byte

var (
	encoding  = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding)
	machineID = readMachineID()
	pid       = uint16(os.Getpid())
	counter   = randomUint32()
)

func readMachineID() [3]byte {
	var id [3]byte
	if host, err := os.Hostname(); err == nil && host != "" {
		sum := md5.Sum([]byte(host))
		copy(id[:], sum[:])
		return id
	}
	if _, err := rand.Read(id[:]); err != nil {
		panic("xid: cannot read random bytes: " + err.Error())
	}
	return id
}

func randomUint32() *atomic.Uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("xid: cannot read random bytes: " + err.Error())
	}
	var c atomic.Uint32
	c.Store(binary.BigEndian.Uint32(b[:]))
	return &c
}

// New returns a new ID for the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID for the given time.
func NewWithTime(t time.Time) ID {
	var id ID
	binary.BigEndian.PutUint32(id[0:4], uint32(t.Unix()))
	copy(id[4:7], machineID[:])
	binary.BigEndian.PutUint16(id[7:9], pid)
	c := counter.Add(1)
	id[9], id[10], id[11] = byte(c>>16), byte(c>>8), byte(c)
	return id
}

// String returns the 20-character text form of the ID.
func (id ID) String() string {
	return encoding.EncodeToString(id[:])
}

// Time returns the timestamp part of the ID.
func (id ID) Time() time.Time {
	return time.Unix(int64(binary.BigEndian.Uint32(id[0:4])), 0)
}
