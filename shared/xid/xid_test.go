package xid

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Known vector from github.com/rs/xid.
func TestStringMatchesRsXid(t *testing.T) {
	raw, err := hex.DecodeString("4d88e15b60f486e428412dc9")
	require.NoError(t, err)
	var id ID
	copy(id[:], raw)
	require.Equal(t, "9m4e2mr0ui3e8a215n4g", id.String())
}

func TestNewIsUniqueAndSortable(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-v]{20}$`)
	seen := make(map[string]struct{})
	prev := ""
	for i := 0; i < 10000; i++ {
		s := New().String()
		require.Regexp(t, re, s)
		_, dup := seen[s]
		require.False(t, dup)
		seen[s] = struct{}{}
		require.GreaterOrEqual(t, s[:7], prev)
		prev = s[:7]
	}
	now := time.Unix(time.Now().Unix(), 0)
	require.Equal(t, now, NewWithTime(now).Time())
}

// Captured on the build box with github.com/rs/xid v1.3.0 (id.Machine()):
// 5a4daa. The derivation below is rs/xid's: md5 of the platform host ID.
func TestMachineIDDerivation(t *testing.T) {
	hid, err := readPlatformMachineID()
	if err != nil || hid == "" {
		hid, err = os.Hostname()
		require.NoError(t, err)
	}
	sum := md5.Sum([]byte(hid))
	require.Equal(t, sum[:3], New().Machine())
	require.Equal(t, uint16(processID()), New().Pid())
}
