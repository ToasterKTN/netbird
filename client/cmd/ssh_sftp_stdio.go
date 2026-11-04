//go:build unix || windows

package cmd

import "io"

// stdio joins stdin and stdout into the stream ServeSFTP expects.
type stdio struct {
	io.Reader
	io.WriteCloser
}
