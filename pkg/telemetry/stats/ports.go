package stats

import (
	"context"
	"time"
)

// records metrics
type StatsRecorder interface {
	AddBytes(n uint64)
	AddSample(metric string, d time.Duration)
}

// per worker StatsRecorder
type StatsRecorderFactory interface {
	NewStatsRecorder() StatsRecorder
}

// hold a snapshot of the StatsCollector
type Snapshot struct {
	TotalRequests uint64
	TotalDuration time.Duration
	TotalBytes    uint64
	StatusCodes   map[int]uint64
	StartTime     time.Time
	Uptime        time.Duration
	AvgLatency    time.Duration
	AvgTLS        time.Duration
}

// provide live snapshots
type Snapshotter interface {
	Snapshot() Snapshot
}

// statsObserver receives periodic stats snapshots from a Broadcaster.
type StatsObserver interface {
	OnSnapshot(Snapshot) // OnSnapshot should be non-blocking.
}

type contextKey struct{}

var RecorderKey = contextKey{}

// retrieves a StatRecorder from the context
func FromContext(ctx context.Context) StatsRecorder {
	if r, ok := ctx.Value(RecorderKey).(StatsRecorder); ok {
		return r
	}
	return nil
}

// injects a StatRecorder into the context
func WithRecorder(ctx context.Context, r StatsRecorder) context.Context {
	return context.WithValue(ctx, RecorderKey, r)
}
