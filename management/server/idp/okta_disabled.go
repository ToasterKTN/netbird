//go:build !idp_okta

package idp

import (
	"fmt"

	"github.com/netbirdio/netbird/management/server/telemetry"
)

func newOktaManager(_ OktaClientConfig, _ telemetry.AppMetrics) (Manager, error) {
	return nil, fmt.Errorf("IdP manager %q (Okta) not compiled in, rebuild with -tags idp_okta", "okta")
}
