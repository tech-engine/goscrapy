package stats

import (
	"context"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type Broadcaster struct {
	snapshotter ISnapshotter
	mu          sync.RWMutex
	observers   []IStatsObserver
	opts        BroadcasterOpts
	logger      core.ILogger
}

type BroadcasterOpts struct {
	Interval time.Duration
}

func defaultOpts() BroadcasterOpts {
	return BroadcasterOpts{
		Interval: 500 * time.Millisecond,
	}
}

func WithInterval(d time.Duration) types.OptFunc[BroadcasterOpts] {
	return func(o *BroadcasterOpts) {
		o.Interval = d
	}
}

func NewBroadcaster(snapshotter ISnapshotter, optFuncs ...types.OptFunc[BroadcasterOpts]) *Broadcaster {
	opts := defaultOpts()
	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &Broadcaster{
		snapshotter: snapshotter,
		opts:        opts,
		observers:   make([]IStatsObserver, 0),
		logger:      logger.EnsureLogger(nil).WithName("StatsBroadcaster"),
	}
}

func (b *Broadcaster) Start(ctx context.Context) error {
	ticker := time.NewTicker(b.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			snapshot := b.snapshotter.Snapshot()

			b.mu.RLock()
			for _, observer := range b.observers {
				observer.OnSnapshot(snapshot)
			}
			b.mu.RUnlock()
		}
	}
}

func (b *Broadcaster) WithLogger(loggerIn core.ILogger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	loggerIn = logger.EnsureLogger(loggerIn)
	b.logger = loggerIn.WithName("StatsBroadcaster")
}

func (b *Broadcaster) Subscribe(observers ...IStatsObserver) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observers = append(b.observers, observers...)
}
