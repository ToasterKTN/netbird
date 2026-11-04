//go:build !rosenpass

package rosenpass

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Manager is a placeholder in builds without the rosenpass tag; it can never
// be constructed, so its methods are unreachable.
type Manager struct{}

// NewManager always fails: rosenpass support is not compiled in.
func NewManager(_ *wgtypes.Key, _ string, _ wgtypes.Key) (*Manager, error) {
	return nil, ErrNotCompiled
}

func (m *Manager) GetPubKey() []byte                                  { return nil }
func (m *Manager) GetAddress() *net.UDPAddr                           { return nil }
func (m *Manager) Run() error                                         { return ErrNotCompiled }
func (m *Manager) Close() error                                       { return nil }
func (m *Manager) SetInterface(PresharedKeySetter)                    {}
func (m *Manager) OnConnected(_ string, _ []byte, _ string, _ string) {}
func (m *Manager) OnDisconnected(string)                              {}
func (m *Manager) IsPresharedKeyInitialized(string) bool              { return false }
