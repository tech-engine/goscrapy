package stats

import (
	"context"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
)

// Coordinates IStatsCollector and IStatsObserver
type TelemetryHub struct {
	mu         sync.RWMutex
	collectors []IStatsCollector
	observers  []IStatsObserver
	opts       TelemetryHubOpts
	startTime  time.Time
}

type TelemetryHubOpts struct {
	Interval time.Duration
}

func defaultOpts() TelemetryHubOpts {
	return TelemetryHubOpts{
		Interval: 500 * time.Millisecond,
	}
}

func WithInterval(d time.Duration) types.OptFunc[TelemetryHubOpts] {
	return func(o *TelemetryHubOpts) {
		o.Interval = d
	}
}

func NewTelemetryHub(optFuncs ...types.OptFunc[TelemetryHubOpts]) *TelemetryHub {
	opts := defaultOpts()
	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &TelemetryHub{
		opts:       opts,
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
	ticker := time.NewTicker(th.opts.Interval)
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

func (th *TelemetryHub) broadcast() {
	th.mu.RLock()
	defer th.mu.RUnlock()

	// If no observers, there is no point in broadcasting
	if len(th.observers) == 0 {
		return
	}

	snap := GlobalSnapshot{
		Timestamp:  time.Now(),
		Uptime:     time.Since(th.startTime),
		Interval:   th.opts.Interval,
		Components: make(map[string]ComponentSnapshot),
	}

	for _, c := range th.collectors {
		snap.Components[c.Name()] = c.Snapshot()
	}

	for _, o := range th.observers {
		o.OnSnapshot(snap)
	}
}
