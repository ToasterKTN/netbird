//go:build js || !upnp

package portforward

import (
	"runtime"

	"context"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// Mapping represents an active NAT port mapping.
type Mapping struct {
	Protocol     string
	InternalPort uint16
	ExternalPort uint16
	ExternalIP   net.IP
	NATType      string
	// TTL is the lease duration. Zero means a permanent lease that never expires.
	TTL time.Duration
}

// Manager is a stub for js/wasm builds (where NAT-PMP/UPnP is not supported)
// and for builds without the upnp tag.
type Manager struct{}

// NewManager returns a stub manager.
func NewManager() *Manager {
	return &Manager{}
}

// Start is a no-op: no NAT port mapping is attempted, as if
// NB_DISABLE_NAT_MAPPER were set.
func (m *Manager) Start(context.Context, uint16) {
	if runtime.GOOS != "js" {
		log.Info("NAT port mapping (UPnP/NAT-PMP/PCP) not compiled in, rebuild with -tags upnp")
	}
}

// GracefullyStop is a no-op.
func (m *Manager) GracefullyStop(context.Context) error { return nil }

// GetMapping always returns nil.
func (m *Manager) GetMapping() *Mapping {
	return nil
}
