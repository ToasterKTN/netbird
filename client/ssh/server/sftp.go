package server

import (
	"errors"

	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

// ErrSFTPNotCompiled is returned when SFTP is used in a build without the sftp tag.
var ErrSFTPNotCompiled = errors.New("SFTP support not compiled in, rebuild with -tags sftp")

// SetAllowSFTP enables or disables SFTP support
func (s *Server) SetAllowSFTP(allow bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allowSFTP = allow
}

// sftpSubsystemHandler handles SFTP subsystem requests
func (s *Server) sftpSubsystemHandler(sess ssh.Session) {
	sessionKey := s.registerSession(sess, cmdSFTP)
	defer s.unregisterSession(sessionKey)

	jwtUsername := s.associateJWTUsername(sess, sessionKey)

	logger := log.WithField("session", sessionKey)
	if jwtUsername != "" {
		logger = logger.WithField("jwt_user", jwtUsername)
	}
	logger.Info("SFTP session started")
	defer logger.Info("SFTP session closed")

	s.mu.RLock()
	allowSFTP := s.allowSFTP
	s.mu.RUnlock()

	if allowSFTP && !SFTPCompiled {
		logger.Warn("SFTP subsystem request denied: " + ErrSFTPNotCompiled.Error())
		allowSFTP = false
	}

	if !allowSFTP {
		logger.Debug("SFTP subsystem request denied: SFTP disabled")
		if err := sess.Exit(1); err != nil {
			logger.Debugf("SFTP session exit: %v", err)
		}
		return
	}

	result := s.CheckPrivileges(PrivilegeCheckRequest{
		RequestedUsername:         sessionUser(sess),
		FeatureSupportsUserSwitch: true,
		FeatureName:               FeatureSFTP,
	})

	if !result.Allowed {
		logger.Warnf("SFTP access denied: %v", result.Error)
		if err := sess.Exit(1); err != nil {
			logger.Debugf("exit SFTP session: %v", err)
		}
		return
	}

	if !result.RequiresUserSwitching {
		if err := s.executeSftpDirect(sess); err != nil {
			logger.Errorf("SFTP direct execution: %v", err)
		}
		return
	}

	if err := s.executeSftpWithPrivilegeDrop(sess, result.User); err != nil {
		logger.Errorf("SFTP privilege drop execution: %v", err)
	}
}

// executeSftpDirect executes SFTP directly without privilege dropping
func (s *Server) executeSftpDirect(sess ssh.Session) error {
	return ServeSFTP(sess)
}
