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
	window := 100 * time.Millisecond

	executor := &mockExecutor{execTime: execTime}

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

	p, err := NewPool(config)
	assert.NoError(t, err)

	// we drain results to prevent blocking
	go func() {
		for range p.Results() {
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	initialWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), initialWorkers)

	// high load burst
	stopBurst := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopBurst:
				return
			default:
				p.Submit(&core.Request{}, "cb", nil)
				time.Sleep(2 * time.Millisecond)
			}
		}
	}()

	time.Sleep(2 * time.Second) // scaling window

	burstWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Greater(t, burstWorkers, int32(minWorkers))
	t.Logf("Scaled up to %d workers", burstWorkers)

	// scaling cooldown
	close(stopBurst)

	// let's wait for EMA decay and worker exit
	time.Sleep(8 * time.Second)

	finalWorkers := p.(*workerPool).activeWorkers.Load()
	assert.Equal(t, int32(minWorkers), finalWorkers)
	t.Logf("Scaled down to %d workers", finalWorkers)
}
