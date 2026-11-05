package version

import (
	"testing"

	v "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeetsMinVersion(t *testing.T) {
	tests := []struct {
		name    string
		minVer  string
		peerVer string
		want    bool
		wantErr bool
	}{
		{
			name:    "Peer version greater than min version",
			minVer:  "0.26.0",
			peerVer: "0.60.1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Peer version equals min version",
			minVer:  "1.0.0",
			peerVer: "1.0.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Peer version less than min version",
			minVer:  "1.0.0",
			peerVer: "0.9.9",
			want:    false,
			wantErr: false,
		},
		{
			name:    "Peer version with pre-release tag greater than min version",
			minVer:  "1.0.0",
			peerVer: "1.0.1-alpha",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Invalid peer version format",
			minVer:  "1.0.0",
			peerVer: "dev",
			want:    false,
			wantErr: true,
		},
		{
			name:    "Invalid min version format",
			minVer:  "invalid.version",
			peerVer: "1.0.0",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MeetsMinVersion(tt.minVer, tt.peerVer)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		candidate string
		want      bool
	}{
		{
			name:      "fork prerelease equals upstream release: no update",
			base:      "0.82.0-ktn.2",
			candidate: "0.82.0",
			want:      false,
		},
		{
			name:      "upstream higher minor: update",
			base:      "0.82.0-ktn.2",
			candidate: "0.83.0",
			want:      true,
		},
		{
			name:      "plain patch newer than fork prerelease: update",
			base:      "0.82.0-ktn.2",
			candidate: "0.82.1",
			want:      true,
		},
		{
			name:      "plain equal versions: no update",
			base:      "0.82.0",
			candidate: "0.82.0",
			want:      false,
		},
		{
			name:      "plain higher patch: update",
			base:      "0.82.0",
			candidate: "0.82.1",
			want:      true,
		},
		{
			name:      "current ahead of candidate: no update",
			base:      "0.83.0",
			candidate: "0.82.0",
			want:      false,
		},
		{
			name:      "fork prerelease vs same numeric with other suffix: no update",
			base:      "0.82.0-ktn.3",
			candidate: "0.82.0-rc.1",
			want:      false,
		},
		{
			name:      "invalid base: no update",
			base:      "development",
			candidate: "0.82.0",
			want:      false,
		},
		{
			name:      "invalid candidate: no update",
			base:      "0.82.0",
			candidate: "not-a-version",
			want:      false,
		},
		{
			name:      "numeric ordering not lexicographic",
			base:      "0.9.0",
			candidate: "0.10.0",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNewer(tt.base, tt.candidate))
		})
	}
}

func TestGreaterThanCore(t *testing.T) {
	must := func(s string) *v.Version {
		ver, err := v.NewVersion(s)
		require.NoError(t, err)
		return ver
	}

	assert.False(t, GreaterThanCore(must("0.82.0"), must("0.82.0-ktn.2")))
	assert.True(t, GreaterThanCore(must("0.83.0"), must("0.82.0-ktn.2")))
	assert.True(t, GreaterThanCore(must("0.82.1"), must("0.82.0-ktn.2")))
	assert.False(t, GreaterThanCore(nil, must("0.82.0")))
	assert.False(t, GreaterThanCore(must("0.82.0"), nil))
}
