package xid

import (
	"encoding/hex"
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
