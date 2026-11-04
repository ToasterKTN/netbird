//go:build !idp_authentik

package idp

import (
	"fmt"

	"github.com/netbirdio/netbird/management/server/telemetry"
)

func newAuthentikManager(_ AuthentikClientConfig, _ telemetry.AppMetrics) (Manager, error) {
	return nil, fmt.Errorf("IdP manager %q (Authentik) not compiled in, rebuild with -tags idp_authentik", "authentik")
}
