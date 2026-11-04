//go:build !idp_google

package idp

import (
	"context"
	"fmt"

	"github.com/netbirdio/netbird/management/server/telemetry"
)

func newGoogleManager(_ context.Context, _ GoogleWorkspaceClientConfig, _ telemetry.AppMetrics) (Manager, error) {
	return nil, fmt.Errorf("IdP manager %q (Google Workspace) not compiled in, rebuild with -tags idp_google", "google")
}
