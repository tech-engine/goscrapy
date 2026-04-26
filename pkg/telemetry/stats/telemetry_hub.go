package stats

import (
	"context"
	"sync"
	"time"
)

type TelemetryHubConfig struct {
	Interval time.Duration
}

// Coordinates IStatsCollector and IStatsObserver
type TelemetryHub struct {
	mu         sync.RWMutex
	collectors []IStatsCollector
	observers  []IStatsObserver
	config     TelemetryHubConfig
	startTime  time.Time
}

func NewTelemetryHub(config *TelemetryHubConfig) *TelemetryHub {
	cfg := TelemetryHubConfig{
		Interval: 500 * time.Millisecond,
	}
	if config != nil {
		if config.Interval > 0 {
			cfg.Interval = config.Interval
		}
	}

	return &TelemetryHub{
		config:     cfg,
		collectors: make([]IStatsCollector, 0),
		observers:  make([]IStatsObserver, 0),
		startTime:  time.Now(),
	}
}

func (th *TelemetryHub) AddCollector(coll IStatsCollector) {
	th.mu.Lock()
	defer th.mu.Unlock()
	th.collectors = append(th.collectors, coll)
}

func (th *TelemetryHub) AddObserver(obs IStatsObserver) {
	th.mu.Lock()
	defer th.mu.Unlock()
	th.observers = append(th.observers, obs)
}

func (th *TelemetryHub) Start(ctx context.Context) error {
	ticker := time.NewTicker(th.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			th.broadcast()
		}
	}
}

// Snapshot returns a view of all registered collectors.
func (th *TelemetryHub) Snapshot() GlobalSnapshot {
	th.mu.RLock()
	defer th.mu.RUnlock()
	return th.snapshotLocked()
}

// must be called with mu.RLock held
func (th *TelemetryHub) snapshotLocked() GlobalSnapshot {
	snap := GlobalSnapshot{
		Timestamp:  time.Now(),
		Uptime:     time.Since(th.startTime),
		Interval:   th.config.Interval,
		Components: make(map[string]ComponentSnapshot),
	}

	for _, c := range th.collectors {
		snap.Components[c.Name()] = c.Snapshot()
	}

	return snap
}

func (th *TelemetryHub) broadcast() {
	th.mu.RLock()
	defer th.mu.RUnlock()

	// If no observers, there is no point in broadcasting
	if len(th.observers) == 0 {
		return
	}

	snap := th.snapshotLocked()

	for _, o := range th.observers {
		o.OnSnapshot(snap)
	}
}
