//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	sshserver "github.com/netbirdio/netbird/client/ssh/server"
)

var (
	sftpWorkingDir  string
	windowsUsername string
	windowsDomain   string
)

var sshSftpCmd = &cobra.Command{
	Use:    "sftp",
	Short:  "SFTP server with user switching for Windows (internal use)",
	Hidden: true,
	RunE:   sftpMain,
}

func init() {
	sshSftpCmd.Flags().StringVar(&sftpWorkingDir, "working-dir", "", "Working directory")
	sshSftpCmd.Flags().StringVar(&windowsUsername, "windows-username", "", "Windows username for user switching")
	sshSftpCmd.Flags().StringVar(&windowsDomain, "windows-domain", "", "Windows domain for user switching")
}

func sftpMain(cmd *cobra.Command, _ []string) error {
	return sftpMainDirect(cmd)
}

func sftpMainDirect(cmd *cobra.Command) error {
	currentUser, err := user.Current()
	if err != nil {
		cmd.PrintErrf("failed to get current user: %v\n", err)
		os.Exit(sshserver.ExitCodeValidationFail)
	}

	if windowsUsername != "" {
		expectedUsername := windowsUsername
		if windowsDomain != "" {
			expectedUsername = fmt.Sprintf(`%s\%s`, windowsDomain, windowsUsername)
		}
		if !strings.EqualFold(currentUser.Username, expectedUsername) && !strings.EqualFold(currentUser.Username, windowsUsername) {
			cmd.PrintErrf("user switching failed\n")
			os.Exit(sshserver.ExitCodeValidationFail)
		}
	}

	log.Debugf("SFTP process running as: %s (UID: %s, Name: %s)", currentUser.Username, currentUser.Uid, currentUser.Name)

	if sftpWorkingDir != "" {
		if err := os.Chdir(sftpWorkingDir); err != nil {
			cmd.PrintErrf("failed to change to working directory %s: %v\n", sftpWorkingDir, err)
		}
	}

	log.Tracef("starting SFTP server")
	if err := sshserver.ServeSFTP(stdio{Reader: os.Stdin, WriteCloser: os.Stdout}); err != nil {
		cmd.PrintErrf("SFTP server error: %v\n", err)
		os.Exit(sshserver.ExitCodeShellExecFail)
	}
	os.Exit(sshserver.ExitCodeSuccess)
	return nil
}
