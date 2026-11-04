//go:build !rosenpass

package rosenpass

import (
	"errors"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestNewManagerNotCompiled(t *testing.T) {
	m, err := NewManager(nil, "wt0", wgtypes.Key{})
	if m != nil || !errors.Is(err, ErrNotCompiled) {
		t.Fatalf("expected ErrNotCompiled, got %v %v", m, err)
	}
}
