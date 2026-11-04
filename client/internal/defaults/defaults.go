// Package defaults holds client default values that lightweight consumers
// (profile config, UI) need without importing the WireGuard interface or
// routing engine packages.
package defaults

import "time"

const (
	// DefaultMTU is the default WireGuard interface MTU.
	DefaultMTU = 1280
	// MinMTU is the minimum allowed WireGuard interface MTU.
	MinMTU = 576
	// MaxMTU is the maximum allowed WireGuard interface MTU.
	MaxMTU = 8192
	// DefaultWgPort is the default WireGuard listen port.
	DefaultWgPort = 51820
	// DNSRouteInterval is the default refresh interval for dynamic (DNS) routes.
	DNSRouteInterval = time.Minute
)
