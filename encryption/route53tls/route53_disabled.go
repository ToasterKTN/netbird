//go:build !route53 && !js

package route53tls

import "crypto/tls"

// Enabled reports whether route53 support is compiled in.
const Enabled = false

// GetCertificate always fails: route53 support is not compiled in.
func (r *Route53TLS) GetCertificate() (*tls.Config, error) {
	return nil, ErrNotCompiled
}
