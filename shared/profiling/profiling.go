// Package profiling starts Pyroscope continuous profiling when configured via
// NB_PYROSCOPE_* env vars. The Pyroscope client is only compiled with the
// "profiling" build tag.
package profiling

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
)

var errNotConfigured = errors.New("pyroscope not configured")

type config struct {
	Address  string // NB_PYROSCOPE_ADDRESS
	User     string // NB_PYROSCOPE_USER, required when Address is set
	Password string // NB_PYROSCOPE_PASSWORD, required when Address is set
}

func loadConfig() (config, error) {
	cfg := config{
		Address:  os.Getenv("NB_PYROSCOPE_ADDRESS"),
		User:     os.Getenv("NB_PYROSCOPE_USER"),
		Password: os.Getenv("NB_PYROSCOPE_PASSWORD"),
	}
	if cfg.Address == "" {
		return cfg, errNotConfigured
	}
	var missing []string
	if cfg.User == "" {
		missing = append(missing, `environment variable "NB_PYROSCOPE_USER" should not be empty`)
	}
	if cfg.Password == "" {
		missing = append(missing, `environment variable "NB_PYROSCOPE_PASSWORD" should not be empty`)
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("failed to parse pyroscope config: %s", strings.Join(missing, "; "))
	}
	if err := validateAddress(cfg.Address); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// validateAddress refuses to send the basic-auth credentials in plaintext to
// anything but a loopback or private endpoint.
func validateAddress(address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid pyroscope address %q: %w", address, err)
	}

	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if isLocalOrPrivate(u.Hostname()) {
			return nil
		}
		return fmt.Errorf("insecure pyroscope address %q: use https for non-local endpoints", address)
	default:
		return fmt.Errorf("pyroscope address %q must use http or https", address)
	}
}

func isLocalOrPrivate(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip, err := netip.ParseAddr(host)
	return err == nil && (ip.IsLoopback() || ip.IsPrivate())
}
