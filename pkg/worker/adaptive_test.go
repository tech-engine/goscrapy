package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type mockExecutor struct {
	execTime time.Duration
}

func (e *mockExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	time.Sleep(e.execTime)
	res.WriteStatusCode(200)
	return nil
}

func TestAdaptiveScaling(t *testing.T) {
	minWorkers := uint32(2)
	maxWorkers := uint32(10)
	execTime := 50 * time.Millisecond
	window := 200 * time.Millisecond

	executor := &mockExecutor{execTime: execTime}

	config := &Config{
		Executor: executor,
		Autoscaler: &AutoscalerConfig{
			MinWorkers:    minWorkers,
			MaxWorkers:    maxWorkers,
			ScalingFactor: 1.5,
			EMAAlpha:      0.9,
			ScalingWindow: window,
		},
	}

	p, err := NewPool(config)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	initialWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), initialWorkers)

	// High Load, should scale Up
	stopBurst := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopBurst:
				return
			default:
				p.Submit(&core.Request{}, "cb", nil)
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Wait for several scaling windows
	time.Sleep(2 * time.Second)

	burstWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Greater(t, burstWorkers, int32(minWorkers))
	t.Logf("Scaled up to %d workers", burstWorkers)

	// No Load, should scale Down
	close(stopBurst)

	// Wait for the EMA to decay and scaling to react
	time.Sleep(3 * time.Second)

	finalWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), finalWorkers)
	t.Logf("Scaled down to %d workers", finalWorkers)
}

// func TestAdaptiveScaling_DisabledByDefault(t *testing.T) {
// 	executor := &mockExecutor{execTime: 10 * time.Millisecond}

// 	s := New(executor, WithWorkers(4))

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go s.Start(ctx)
// 	time.Sleep(100 * time.Millisecond)

// 	// Even with load, worker count should stay fixed
// 	for i := 0; i < 100; i++ {
// 		req := s.NewRequest(ctx)
// 		s.Schedule(req, func(ctx context.Context, res core.IResponseReader) {})
// 		time.Sleep(2 * time.Millisecond)
// 	}
// 	time.Sleep(500 * time.Millisecond)

// 	workers := s.currentWorkerCnt.Load()
// 	if workers != 4 {
// 		t.Errorf("expected fixed worker count 4, got %d (adaptive should be disabled)", workers)
// 	}
// }

// func TestAdaptiveScaling_Snapshot(t *testing.T) {
// 	executor := &mockExecutor{execTime: 10 * time.Millisecond}

// 	s := New(executor,
// 		WithWorkers(4),
// 		WithAdaptiveScaling(AdaptiveScalingConfig{
// 			MinWorkers: 4,
// 			MaxWorkers: 20,
// 		}),
// 	)

// 	// Before start, snapshot should be reasonable
// 	snap := s.Snapshot().(SchedulerSnapshot)
// 	if snap.CurrentWorkerCnt != 0 {
// 		t.Errorf("expected 0 workers before start, got %d", snap.CurrentWorkerCnt)
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go s.Start(ctx)
// 	time.Sleep(100 * time.Millisecond)

// 	snap = s.Snapshot().(SchedulerSnapshot)
// 	if snap.CurrentWorkerCnt != 4 {
// 		t.Errorf("expected 4 workers after start, got %d", snap.CurrentWorkerCnt)
// 	}

// 	// Name should be consistent
// 	if s.Name() != "Scheduler" {
// 		t.Errorf("expected Name() = 'Scheduler', got '%s'", s.Name())
// 	}
// }
