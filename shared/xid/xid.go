// Package xid generates globally unique, time-sortable 20-character IDs
// identical in layout and derivation to github.com/rs/xid v1.3.0: 4-byte
// big-endian Unix seconds, first 3 bytes of md5(platform host ID, else
// hostname), 2-byte PID (xor crc32 of /proc/self/cpuset inside containers),
// and a 3-byte counter seeded randomly, encoded as lowercase base32hex.
//
// Ported from github.com/rs/xid, Copyright (c) 2015 Olivier Poitrey,
// MIT License (see LICENSE in this directory).
package xid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync/atomic"
	"time"
)

// ID is a 12-byte unique identifier.
type ID [12]byte

var (
	encoding        = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding)
	objectIDCounter = randInt()
	machineID       = readMachineID()
	pid             = processID()
)

func processID() int {
	p := os.Getpid()
	// Inside a container /proc/self/cpuset is not "/"; mix it in as rs/xid does.
	if b, err := os.ReadFile("/proc/self/cpuset"); err == nil && len(b) > 1 {
		p ^= int(crc32.ChecksumIEEE(b))
	}
	return p
}

func readMachineID() []byte {
	id := make([]byte, 3)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err == nil && len(hid) != 0 {
		sum := md5.Sum([]byte(hid))
		copy(id, sum[:])
	} else if _, randErr := rand.Read(id); randErr != nil {
		panic(fmt.Errorf("xid: cannot get hostname nor generate a random number: %v; %v", err, randErr))
	}
	return id
}

func randInt() uint32 {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("xid: cannot generate random number: %v;", err))
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// New returns a new ID for the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID for the given time.
func NewWithTime(t time.Time) ID {
	var id ID
	binary.BigEndian.PutUint32(id[:], uint32(t.Unix()))
	id[4], id[5], id[6] = machineID[0], machineID[1], machineID[2]
	id[7], id[8] = byte(pid>>8), byte(pid)
	i := atomic.AddUint32(&objectIDCounter, 1)
	id[9], id[10], id[11] = byte(i>>16), byte(i>>8), byte(i)
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

// Machine returns the 3-byte machine ID part.
func (id ID) Machine() []byte { return id[4:7] }

// Pid returns the process ID part.
func (id ID) Pid() uint16 { return binary.BigEndian.Uint16(id[7:9]) }
