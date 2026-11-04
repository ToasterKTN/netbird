package localmetrics

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/netbirdio/netbird/client/internal/metrics"
	"github.com/netbirdio/netbird/client/internal/peer"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/metrics.golden")

const goldenContentType = "text/plain; version=0.0.4; charset=utf-8; escaping=underscores"

// goldenStatus extends testStatus with label values that need escaping.
func goldenStatus() *stubStatus {
	s := testStatus()
	s.signal.Connected = false
	s.peers = append(s.peers,
		peer.State{FQDN: "quote\"back\\slash\nnewline.netbird.cloud", ConnStatus: peer.StatusConnected, Latency: 1500 * time.Microsecond},
		peer.State{FQDN: "ümlaut.netbird.cloud", ConnStatus: peer.StatusConnected, Latency: 3 * time.Second},
	)
	return s
}

// goldenClientMetrics records a fixed set of observations, including values
// exactly on bucket boundaries, above the largest bucket and zero.
func goldenClientMetrics() *metrics.ClientMetrics {
	ctx := context.Background()
	cm := metrics.NewClientMetrics(metrics.AgentInfo{Version: "test", OS: "linux"})
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stages := func(sig, ready, wg time.Duration) metrics.ConnectionStageTimestamps {
		return metrics.ConnectionStageTimestamps{SignalingReceived: base.Add(sig), ConnectionReady: base.Add(ready), WgHandshakeSuccess: base.Add(wg)}
	}
	cm.RecordConnectionStages(ctx, "peerkey-a", metrics.ConnectionTypeICEP2P, false, stages(0, 100*time.Millisecond, 350*time.Millisecond))
	cm.RecordConnectionStages(ctx, "peerkey-b", metrics.ConnectionTypeRelay, true, stages(0, 2*time.Second, 90*time.Second))
	cm.RecordConnectionStages(ctx, "peerkey-c", metrics.ConnectionTypeICETurn, false, stages(0, 50*time.Millisecond, 50*time.Millisecond+333*time.Microsecond))
	cm.RecordSyncDuration(ctx, 5*time.Millisecond)
	cm.RecordSyncDuration(ctx, 1234567*time.Microsecond)
	cm.RecordSyncDuration(ctx, 0)
	cm.RecordSyncPhase(ctx, "network_map", 25*time.Millisecond)
	cm.RecordSyncPhase(ctx, "dns", 7*time.Millisecond)
	cm.RecordSyncPhase(ctx, "a\"b", 11*time.Second)
	cm.RecordLoginDuration(ctx, 800*time.Millisecond, true)
	cm.RecordLoginDuration(ctx, 10*time.Second, false)
	return cm
}

// TestMetricsEndpointGolden pins the exact /metrics output so that changes to
// the exposition implementation cannot silently alter what scrapers see.
func TestMetricsEndpointGolden(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cm := goldenClientMetrics()
	m := NewManager(ctx, goldenStatus(), cm.PrometheusGatherer)
	m.Reconcile(true, addr)

	var body, contentType string
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://%s/metrics", addr))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil || resp.StatusCode != http.StatusOK {
			return false
		}
		body, contentType = string(data), resp.Header.Get("Content-Type")
		return true
	}, 2*time.Second, 50*time.Millisecond)

	require.Equal(t, goldenContentType, contentType)

	path := filepath.Join("testdata", "metrics.golden")
	if *updateGolden {
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), body)
}
