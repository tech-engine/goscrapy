package stats

import (
	"context"
	"time"
)

// StatRecorder records metrics
type StatRecorder interface {
	AddBytes(n uint64)
	AddSample(metric string, d time.Duration)
}

// per worker StatRecorder
type CollectorProducer interface {
	NewWorkerCollector() StatRecorder
}

type contextKey struct{}

var RecorderKey = contextKey{}

// retrieves a StatRecorder from the context
func FromContext(ctx context.Context) StatRecorder {
	if r, ok := ctx.Value(RecorderKey).(StatRecorder); ok {
		return r
	}
	return nil
}

// injects a StatRecorder into the context
func WithRecorder(ctx context.Context, r StatRecorder) context.Context {
	return context.WithValue(ctx, RecorderKey, r)
}
