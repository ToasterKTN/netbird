// Package rosenpass integrates post-quantum Rosenpass key exchange. The
// implementation is only compiled with the "rosenpass" build tag; without it
// NewManager returns ErrNotCompiled. Seed/ratchet helpers are always built.
package rosenpass

import (
	"errors"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ErrNotCompiled is returned by NewManager when built without the rosenpass tag.
var ErrNotCompiled = errors.New("rosenpass support not compiled in, rebuild with -tags rosenpass")

// PresharedKeySetter is the interface for setting preshared keys on WireGuard peers.
// This minimal interface allows rosenpass to update PSKs without depending on the full WGIface.
type PresharedKeySetter interface {
	SetPresharedKey(peerKey string, psk wgtypes.Key, updateOnly bool) error
}
