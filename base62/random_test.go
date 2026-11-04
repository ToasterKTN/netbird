package base62

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"testing"
)

// seqReader returns a deterministic byte stream derived from seed.
func seqReader(seed string) *bytes.Reader {
	var out []byte
	h := sha256.Sum256([]byte(seed))
	for i := 0; i < 64; i++ {
		h = sha256.Sum256(h[:])
		out = append(out, h[:]...)
	}
	return bytes.NewReader(out)
}

// Expected values were produced by github.com/hashicorp/go-secure-stdlib/base62.RandomWithReader
// with the same input streams, to guarantee identical output after replacing that library.
func TestRandomWithReaderMatchesLegacyLibrary(t *testing.T) {
	cases := []struct {
		seed   string
		length int
		want   string
	}{
		{"a", 1, "F"},
		{"b", 1, "5"},
		{"a", 6, "Ff67Au"},
		{"b", 6, "52RiU8"},
		{"a", 30, "Ff67AuMJsZRtfz1zlKnRVsSnbHG4ex"},
		{"b", 30, "52RiU8oZMENzZNDQBAfsffZEFCS5U0"},
		{"a", 32, "Ff67AuMJsZRtfz1zlKnRVsSnbHG4exKD"},
		{"b", 32, "52RiU8oZMENzZNDQBAfsffZEFCS5U0R9"},
		{"a", 100, "Ff67AuMJsZRtfz1zlKnRVsSnbHG4exKDVFhw6v10I9wrjsPNbFXgScMU5LDdUXhFFQfeWG4Mw9NEZw3H4biBkr2j8o22dCFH45ef"},
		{"b", 100, "52RiU8oZMENzZNDQBAfsffZEFCS5U0R9qAsaIqMnaue2HsywAkultu7HPqYvXHedeJRVPHuWfpbtwjXelWwLORQ3uZ5F6Y71RlIC"},
	}
	for _, tc := range cases {
		got, err := RandomWithReader(tc.length, seqReader(tc.seed))
		if err != nil {
			t.Fatalf("seed %q len %d: %v", tc.seed, tc.length, err)
		}
		if got != tc.want {
			t.Errorf("seed %q len %d: got %q, want %q", tc.seed, tc.length, got, tc.want)
		}
	}
}

func TestRandom(t *testing.T) {
	if s, err := Random(0); err != nil || s != "" {
		t.Fatalf("Random(0) = %q, %v", s, err)
	}
	s, err := Random(40)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 40 {
		t.Fatalf("len = %d", len(s))
	}
	for _, c := range s {
		if !strings.ContainsRune(randomCharset, c) {
			t.Fatalf("unexpected char %q", c)
		}
	}
}

func TestRandomWithReaderShortRead(t *testing.T) {
	if _, err := RandomWithReader(10, bytes.NewReader([]byte{1, 2})); err == nil {
		t.Fatal("expected error on short read")
	}
}
