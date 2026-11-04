//go:build !js

// Package route53tls obtains TLS certificates via ACME DNS-01 on AWS Route53.
// The certmagic/AWS implementation is only compiled with the "route53" build
// tag; without it GetCertificate returns ErrNotCompiled.
package route53tls

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotCompiled is returned when the binary was built without the route53 tag.
var ErrNotCompiled = errors.New("route53 (Let's Encrypt DNS-01) support not compiled in, rebuild with -tags route53")

// Route53TLS by default, loads the AWS configuration from the environment.
// env variables: AWS_REGION, AWS_PROFILE, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
type Route53TLS struct {
	DataDir string
	Email   string
	Domains []string
	CA      string
}

func emailFromDomain(domain string) string {
	if domain == "" {
		return ""
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	if parts[0] == "" {
		return ""
	}
	return fmt.Sprintf("admin@%s.%s", parts[len(parts)-2], parts[len(parts)-1])
}
