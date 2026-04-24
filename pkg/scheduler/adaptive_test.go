// will revist later on free time
package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/tech-engine/goscrapy/internal/request"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type mockExecutor struct {
	execTime time.Duration
}

func (e *mockExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	time.Sleep(e.execTime)
	return nil
}

func (e *mockExecutor) WithLogger(core.ILogger) IExecutor {
	return e
}

func TestAdaptiveScaling(t *testing.T) {
	minWorkers := uint16(2)
	maxWorkers := uint16(10)
	execTime := 50 * time.Millisecond
	window := 200 * time.Millisecond

	executor := &mockExecutor{execTime: execTime}

	s := New(executor, request.NewPool(),
		WithWorkers(minWorkers),
		WithAdaptiveScaling(AdaptiveScalingConfig{
			MinWorkers:    minWorkers,
			MaxWorkers:    maxWorkers,
			ScalingFactor: 1.5,
			EMAAlpha:      0.9,
			ScalingWindow: window,
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	initialWorkers := s.currentWorkerCnt.Load()
	if initialWorkers != int32(minWorkers) {
		t.Fatalf("expected initial workers %d, got %d", minWorkers, initialWorkers)
	}

	// Phase 1: High Load -> Scale Up
	stopBurst := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopBurst:
				return
			default:
				req := request.NewPool().Acquire(ctx)
				s.Schedule(req, func(ctx context.Context, res core.IResponseReader) {})
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Wait for several scaling windows
	time.Sleep(2 * time.Second)

	burstWorkers := s.currentWorkerCnt.Load()
	if burstWorkers <= int32(minWorkers) {
		t.Errorf("expected workers to scale up from %d, stayed at %d", minWorkers, burstWorkers)
	}
	t.Logf("Scaled up to %d workers", burstWorkers)

	// Phase 2: No Load -> Scale Down
	close(stopBurst)

	// Wait for the EMA to decay and scaling to react
	time.Sleep(3 * time.Second)

	finalWorkers := s.currentWorkerCnt.Load()
	if finalWorkers > int32(minWorkers) {
		t.Errorf("expected workers to scale down to %d, stayed at %d", minWorkers, finalWorkers)
	}
	t.Logf("Scaled down to %d workers", finalWorkers)
}

func TestAdaptiveScaling_DisabledByDefault(t *testing.T) {
	executor := &mockExecutor{execTime: 10 * time.Millisecond}

	s := New(executor, request.NewPool(), WithWorkers(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Even with load, worker count should stay fixed
	for i := 0; i < 100; i++ {
		req := request.NewPool().Acquire(ctx)
		s.Schedule(req, func(ctx context.Context, res core.IResponseReader) {})
		time.Sleep(2 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)

	workers := s.currentWorkerCnt.Load()
	if workers != 4 {
		t.Errorf("expected fixed worker count 4, got %d (adaptive should be disabled)", workers)
	}
}

func TestAdaptiveScaling_Snapshot(t *testing.T) {
	executor := &mockExecutor{execTime: 10 * time.Millisecond}

	s := New(executor, request.NewPool(),
		WithWorkers(4),
		WithAdaptiveScaling(AdaptiveScalingConfig{
			MinWorkers: 4,
			MaxWorkers: 20,
		}),
	)

	// Before start, snapshot should be reasonable
	snap := s.Snapshot().(SchedulerSnapshot)
	if snap.CurrentWorkerCnt != 0 {
		t.Errorf("expected 0 workers before start, got %d", snap.CurrentWorkerCnt)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	snap = s.Snapshot().(SchedulerSnapshot)
	if snap.CurrentWorkerCnt != 4 {
		t.Errorf("expected 4 workers after start, got %d", snap.CurrentWorkerCnt)
	}

	// Name should be consistent
	if s.Name() != "Scheduler" {
		t.Errorf("expected Name() = 'Scheduler', got '%s'", s.Name())
	}
}
