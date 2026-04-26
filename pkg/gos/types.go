package gos

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
	"github.com/tech-engine/goscrapy/pkg/worker"
)

type app[OUT any] struct {
	*core.Core[OUT]
	Engine            core.IEngine[OUT]
	PipelineManager   engine.IPipelineManager[OUT]
	Scheduler         engine.IScheduler
	WorkerPool        engine.IWorkerPool
	Executor          worker.IExecutor
	MiddlewareManager IMiddlewareManager
	httpClient        *http.Client
	logger            core.ILogger
	hub               *ts.TelemetryHub
	cancelableSignal  *cancelableSignal
	lastErr           error
}

type cancelableSignal struct {
	cancel context.CancelFunc
	ctx    context.Context
}

func newCancelableSignal(ctx context.Context) *cancelableSignal {
	ctx, cancel := context.WithCancel(ctx)
	return &cancelableSignal{
		cancel: cancel,
		ctx:    ctx,
	}
}
