//go:build !idp_okta || !idp_authentik

package idp

import (
	"context"
	"strings"
	"testing"
)

func TestNewManagerReportsMissingBuildTag(t *testing.T) {
	for typ, tag := range map[string]string{"okta": "idp_okta", "authentik": "idp_authentik"} {
		_, err := NewManager(context.Background(), Config{ManagerType: typ, ClientConfig: &ClientConfig{}, ExtraConfig: ExtraConfig{}}, nil)
		if err == nil || !strings.Contains(err.Error(), "not compiled in") {
			continue // compiled in for this build
		}
		if !strings.Contains(err.Error(), "-tags "+tag) {
			t.Fatalf("%s: unexpected error %v", typ, err)
		}
	}
}
