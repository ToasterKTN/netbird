// Package localmetrics exposes client connection state as a local
// Prometheus /metrics endpoint.
package localmetrics

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/netbirdio/netbird/client/internal/metrics/promtext"
	"github.com/netbirdio/netbird/client/internal/peer"
)

// DefaultListenAddress is used when local metrics are enabled without an explicit address.
const DefaultListenAddress = "127.0.0.1:9191"

const (
	shutdownTimeout   = 3 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = time.Minute
)

// statusSource provides the connection state snapshots the collector reads on scrape.
type statusSource interface {
	GetPeerStates() []peer.State
	GetManagementState() peer.ManagementState
	GetSignalState() peer.SignalState
}

// GathererProvider returns the current client metrics gatherer, or nil when
// no engine is running. It is called on every scrape.
type GathererProvider func() promtext.Gatherer

// Manager runs the local /metrics HTTP endpoint according to the active
// client configuration. Reconcile is safe to call on every config change.
type Manager struct {
	status        statusSource
	clientMetrics GathererProvider

	mu   sync.Mutex
	srv  *http.Server
	addr string
}

// NewManager creates a manager that serves metrics from status and
// clientMetrics and shuts down when ctx is canceled.
func NewManager(ctx context.Context, status statusSource, clientMetrics GathererProvider) *Manager {
	m := &Manager{status: status, clientMetrics: clientMetrics}
	go func() {
		<-ctx.Done()
		m.Stop()
	}()
	return m
}

// Reconcile starts, stops, or restarts the metrics endpoint to match the
// desired state. An empty addr falls back to DefaultListenAddress.
func (m *Manager) Reconcile(enabled bool, addr string) {
	if addr == "" {
		addr = DefaultListenAddress
	}
	warnIfNotLoopback(addr)

	m.mu.Lock()
	defer m.mu.Unlock()

	if !enabled {
		m.stop()
		return
	}
	if m.srv != nil && m.addr == addr {
		return
	}
	m.stop()

	gatherers := promtext.Gatherers{newCollector(m.status), promtext.GathererFunc(func() []promtext.Family {
		if m.clientMetrics == nil {
			return nil
		}
		g := m.clientMetrics()
		if g == nil {
			return nil
		}
		return g.Gather()
	})}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promtext.Handler(gatherers))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	m.srv = srv
	m.addr = addr

	log.Infof("serving local metrics on http://%s/metrics", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to serve local metrics on %s: %v", addr, err)
			m.clear(srv)
		}
	}()
}

// clear drops the reference to srv so a later Reconcile with the same
// address restarts it. A newer server may already have replaced it, in
// which case the reference must stay.
func (m *Manager) clear(srv *http.Server) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.srv != srv {
		return
	}
	m.srv = nil
	m.addr = ""
}

// Stop shuts down the metrics endpoint if it is running.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stop()
}

// stop shuts down the running server. Callers must hold m.mu.
func (m *Manager) stop() {
	if m.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := m.srv.Shutdown(ctx); err != nil {
		log.Debugf("failed to shut down local metrics server: %v", err)
	}
	m.srv = nil
	m.addr = ""
}

// collector converts status recorder snapshots into metric families at scrape time.
type collector struct {
	status statusSource
}

func newCollector(status statusSource) *collector {
	return &collector{status: status}
}

func gauge(name, help string, metrics ...promtext.Metric) promtext.Family {
	return promtext.Family{Name: name, Help: help, Type: promtext.Gauge, Metrics: metrics}
}

// Gather implements promtext.Gatherer.
func (c *collector) Gather() []promtext.Family {
	peers := c.status.GetPeerStates()

	var p2p, relayed float64
	var latencies []promtext.Metric
	for _, p := range peers {
		if p.ConnStatus != peer.StatusConnected {
			continue
		}
		if p.Relayed {
			relayed++
			continue
		}
		p2p++

		if latency := p.Latency.Seconds(); latency > 0 {
			latencies = append(latencies, promtext.Metric{Labels: []promtext.Label{{Name: "peer", Value: p.FQDN}}, Value: latency})
		}
	}

	return []promtext.Family{
		gauge("netbird_management_connected",
			"Whether the client is connected to the management service (1 connected, 0 disconnected).",
			promtext.Metric{Value: boolToFloat(c.status.GetManagementState().Connected)}),
		gauge("netbird_signal_connected",
			"Whether the client is connected to the signal service (1 connected, 0 disconnected).",
			promtext.Metric{Value: boolToFloat(c.status.GetSignalState().Connected)}),
		gauge("netbird_peers",
			"Number of peers known to this client.",
			promtext.Metric{Value: float64(len(peers))}),
		gauge("netbird_peers_connected",
			"Number of connected peers by connection type.",
			promtext.Metric{Labels: []promtext.Label{{Name: "connection_type", Value: "p2p"}}, Value: p2p},
			promtext.Metric{Labels: []promtext.Label{{Name: "connection_type", Value: "relay"}}, Value: relayed}),
		gauge("netbird_peer_latency_seconds",
			"Round-trip latency per directly connected peer; relayed connections have no latency measurement.",
			latencies...),
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// IsLoopback reports whether addr binds the endpoint to the local host only.
// An empty address means DefaultListenAddress. It fails closed: an address
// that cannot be confirmed loopback, including an unparseable one, is not.
func IsLoopback(addr string) bool {
	if addr == "" {
		addr = DefaultListenAddress
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return ip.Unmap().IsLoopback()
}

// warnIfNotLoopback logs a warning when the listen address cannot be
// confirmed to be local-only, since the endpoint exposes peer and
// connectivity details without authentication.
func warnIfNotLoopback(addr string) {
	if IsLoopback(addr) {
		return
	}
	log.Warnf("local metrics endpoint listens on non-loopback address %s and is reachable from the network without authentication", addr)
}
