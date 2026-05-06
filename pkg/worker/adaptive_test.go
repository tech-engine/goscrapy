package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type adaptiveTestExecutor struct {
	execTime time.Duration
}

func (e *adaptiveTestExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	time.Sleep(e.execTime)
	res.WriteStatusCode(200)
	return nil
}

func TestAdaptive_Scaling(t *testing.T) {
	minWorkers := uint32(2)
	maxWorkers := uint32(10)
	execTime := 50 * time.Millisecond
	window := 100 * time.Millisecond

	executor := &adaptiveTestExecutor{execTime: execTime}

	config := &Config{
		Executor: executor,
		Autoscaler: &AutoscalerConfig{
			MinWorkers:    minWorkers,
			MaxWorkers:    maxWorkers,
			ScalingFactor: 2.0, // aggressive scaling for test
			EMAAlpha:      0.5,
			ScalingWindow: window,
		},
	}

	wp, err := NewPool(config)
	assert.NoError(t, err)

	// we drain results to prevent blocking
	go func() {
		for range wp.Results() {
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wp.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	initialWorkers := wp.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), initialWorkers)

	// high load burst
	stopBurst := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopBurst:
				return
			default:
				wp.Submit(context.Background(), &core.Request{}, "cb", nil)
				time.Sleep(2 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second) // scaling window

	burstWorkers := wp.(*workerPool).activeWorkers.Load()
	assert.Greater(t, burstWorkers, int32(minWorkers))
	t.Logf("Scaled up to %d workers", burstWorkers)

	// scaling cooldown
	close(stopBurst)

	// let's wait for EMA decay and worker exit
	time.Sleep(8 * time.Second)

	finalWorkers := wp.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), finalWorkers)
	t.Logf("Scaled down to %d workers", finalWorkers)
}
