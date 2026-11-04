//go:build !profiling

package profiling

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// Start only reports configuration: the Pyroscope client is not compiled in.
func Start(applicationName string) func() {
	_, err := loadConfig()
	switch {
	case errors.Is(err, errNotConfigured):
	case err != nil:
		log.Errorf("failed to load pyroscope config: %v", err)
	default:
		log.Warnf("NB_PYROSCOPE_ADDRESS is set but continuous profiling (pyroscope) is not compiled in, rebuild with -tags profiling (%s)", applicationName)
	}
	return func() {}
}
