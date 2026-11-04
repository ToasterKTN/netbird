package base62

import (
	"crypto/rand"
	"fmt"
	"io"
)

// randomCharset is the alphabet used for random secrets. It intentionally differs from
// the checksum alphabet above and matches github.com/hashicorp/go-secure-stdlib/base62,
// which this replaces, so generated tokens keep the same format.
const (
	randomCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomCsLen   = byte(len(randomCharset))
)

// Random returns a cryptographically random base62 string of the given length.
func Random(length int) (string, error) {
	return RandomWithReader(length, rand.Reader)
}

// RandomWithReader returns a random base62 string of the given length using bytes from reader.
// Bytes >= 248 are rejected so every character is uniformly distributed.
func RandomWithReader(length int, reader io.Reader) (string, error) {
	if length <= 0 {
		return "", nil
	}
	output := make([]byte, 0, length)
	buf := make([]byte, length+length/4)
	for {
		if _, err := io.ReadFull(reader, buf); err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}
		for _, b := range buf {
			if b < randomCsLen*4 {
				output = append(output, randomCharset[b%randomCsLen])
				if len(output) == length {
					return string(output), nil
				}
			}
		}
	}
}
