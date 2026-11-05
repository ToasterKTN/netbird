package version

import (
	"strings"

	v "github.com/hashicorp/go-version"
)

// sanitizeVersion removes anything after the pre-release tag (e.g., "-dev", "-alpha", etc.)
func sanitizeVersion(version string) string {
	parts := strings.Split(version, "-")
	return parts[0]
}

// MeetsMinVersion checks if the peer's version meets or exceeds the minimum required version
func MeetsMinVersion(minVer, peerVer string) (bool, error) {
	peerVer = sanitizeVersion(peerVer)
	minVer = sanitizeVersion(minVer)

	peerNBVer, err := v.NewVersion(peerVer)
	if err != nil {
		return false, err
	}

	constraints, err := v.NewConstraint(">= " + minVer)
	if err != nil {
		return false, err
	}

	return constraints.Check(peerNBVer), nil
}

// IsNewer reports whether candidate's numeric version (MAJOR.MINOR.PATCH) is
// strictly greater than base's. Prerelease and metadata suffixes are ignored,
// so "0.82.0" is not newer than "0.82.0-ktn.2". Plain versions keep normal
// numeric ordering (0.83.0 is newer than 0.82.0; 0.82.1 is newer than 0.82.0).
func IsNewer(base, candidate string) bool {
	baseVer, err := v.NewVersion(sanitizeVersion(base))
	if err != nil {
		return false
	}
	candidateVer, err := v.NewVersion(sanitizeVersion(candidate))
	if err != nil {
		return false
	}
	return candidateVer.GreaterThan(baseVer)
}

// GreaterThanCore reports whether latest's MAJOR.MINOR.PATCH is strictly
// greater than current's, ignoring prerelease/metadata on either side.
func GreaterThanCore(latest, current *v.Version) bool {
	if latest == nil || current == nil {
		return false
	}
	return latest.Core().GreaterThan(current.Core())
}
