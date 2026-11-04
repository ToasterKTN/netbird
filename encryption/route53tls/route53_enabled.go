//go:build route53 && !js

package route53tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/route53"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/acme"
)

// Enabled reports whether route53 support is compiled in.
const Enabled = true

func (r *Route53TLS) GetCertificate() (*tls.Config, error) {
	if len(r.Domains) == 0 {
		return nil, fmt.Errorf("no domains provided")
	}

	certmagic.Default.Logger = logger()
	certmagic.Default.Storage = &certmagic.FileStorage{Path: r.DataDir}
	certmagic.DefaultACME.Agreed = true
	if r.Email != "" {
		certmagic.DefaultACME.Email = r.Email
	} else {
		certmagic.DefaultACME.Email = emailFromDomain(r.Domains[0])
	}

	if r.CA == "" {
		certmagic.DefaultACME.CA = certmagic.LetsEncryptProductionCA
	} else {
		certmagic.DefaultACME.CA = r.CA
	}

	certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: &route53.Provider{},
		},
	}
	cm := certmagic.NewDefault()
	if err := cm.ManageSync(context.Background(), r.Domains); err != nil {
		log.Errorf("failed to manage certificate: %v", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1", acme.ALPNProto},
	}

	return tlsConfig, nil
}

func logger() *zap.Logger {
	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		os.Stderr,
		zap.ErrorLevel,
	))
}
