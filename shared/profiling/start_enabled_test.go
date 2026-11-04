//go:build profiling

package profiling

import (
	"testing"

	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartSkipsSecondProfilerInProcess(t *testing.T) {
	clearEnv(t)
	t.Setenv("NB_PYROSCOPE_ADDRESS", "http://127.0.0.1:1")
	t.Setenv("NB_PYROSCOPE_USER", "user")
	t.Setenv("NB_PYROSCOPE_PASSWORD", "token")

	started.Store(true)
	t.Cleanup(func() { started.Store(false) })
	hook := logtest.NewGlobal()
	t.Cleanup(hook.Reset)

	stop := Start("netbird-second")
	stop()

	assert.True(t, started.Load(), "the running profiler must stay marked as started")
	entry := hook.LastEntry()
	require.NotNil(t, entry, "the skipped start must be logged")
	assert.Equal(t, log.WarnLevel, entry.Level)
	assert.Contains(t, entry.Message, "already running")
}
