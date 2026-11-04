//go:build !sftp

package server

import "io"

// SFTPCompiled reports whether the SFTP subsystem is compiled in.
const SFTPCompiled = false

// ServeSFTP always fails: SFTP support is not compiled in.
func ServeSFTP(io.ReadWriteCloser) error {
	return ErrSFTPNotCompiled
}
