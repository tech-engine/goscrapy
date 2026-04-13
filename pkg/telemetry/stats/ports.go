package stats

import (
	"context"
	"time"
)

// Represents a component that records metrics
type IStatsRecorder interface {
	AddBytes(n uint64)
	AddSample(metric string, d time.Duration)
}

// Represents a factory that create IStatsRecorder
type IStatsRecorderFactory interface {
	NewStatsRecorder() IStatsRecorder
}

// Represents a generic interface for individual components to return their state.
type ComponentSnapshot any

// GlobalSnapshot is broadcasted to all observers.
type GlobalSnapshot struct {
	Timestamp  time.Time
	Uptime     time.Duration
	Interval   time.Duration
	Components map[string]ComponentSnapshot
}

// Represents a component that generates stats/matrics.
type IStatsCollector interface {
	Name() string
	Snapshot() ComponentSnapshot
}

// Represnts a component that receives periodic GlobalSnapshots from a Hub.
type IStatsObserver interface {
	OnSnapshot(GlobalSnapshot) // should be non blocking.
}

type contextKey struct{}

var RecorderKey = contextKey{}

// Retrieves a StatRecorder from a passed context
func FromContext(ctx context.Context) IStatsRecorder {
	if r, ok := ctx.Value(RecorderKey).(IStatsRecorder); ok {
		return r
	}
	return nil
}

// Injects a StatRecorder into a passed context
func WithRecorder(ctx context.Context, r IStatsRecorder) context.Context {
	return context.WithValue(ctx, RecorderKey, r)
}
