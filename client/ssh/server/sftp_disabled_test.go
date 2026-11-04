//go:build !sftp

package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServeSFTPNotCompiled(t *testing.T) {
	require.False(t, SFTPCompiled)
	require.ErrorIs(t, ServeSFTP(nil), ErrSFTPNotCompiled)
}
