//go:build sftp

package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
)

// SFTPCompiled reports whether the SFTP subsystem is compiled in.
const SFTPCompiled = true

// ServeSFTP serves the SFTP protocol on rw until the client disconnects.
func ServeSFTP(rw io.ReadWriteCloser) error {
	sftpServer, err := sftp.NewServer(rw)
	if err != nil {
		return fmt.Errorf("SFTP server creation: %w", err)
	}

	defer func() {
		if err := sftpServer.Close(); err != nil {
			log.Debugf("failed to close sftp server: %v", err)
		}
	}()

	if err := sftpServer.Serve(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}
