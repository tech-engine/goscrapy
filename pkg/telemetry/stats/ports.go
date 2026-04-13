package stats

import (
	"context"
	"time"
)

// records metrics
type IStatsRecorder interface {
	AddBytes(n uint64)
	AddSample(metric string, d time.Duration)
}

// per worker StatsRecorder
type IStatsRecorderFactory interface {
	NewStatsRecorder() IStatsRecorder
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
type ISnapshotter interface {
	Snapshot() Snapshot
}

// statsObserver receives periodic stats snapshots from a Broadcaster.
type IStatsObserver interface {
	OnSnapshot(Snapshot) // OnSnapshot should be non-blocking.
}

type contextKey struct{}

var RecorderKey = contextKey{}

// retrieves a StatRecorder from the context
func FromContext(ctx context.Context) IStatsRecorder {
	if r, ok := ctx.Value(RecorderKey).(IStatsRecorder); ok {
		return r
	}
	return nil
}

// injects a StatRecorder into the context
func WithRecorder(ctx context.Context, r IStatsRecorder) context.Context {
	return context.WithValue(ctx, RecorderKey, r)
}
