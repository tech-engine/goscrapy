package scheduler

import (
	"context"
	"net/http"
	"testing"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

type dummyExecutor struct {
}

func (e *dummyExecutor) Execute(reader core.IRequestReader, writer engine.IResponseWriter) error {
	return nil
}

func TestWorker(t *testing.T) {
	// create a worker
	var workerId uint16 = 1
	var respPoolSize uint64 = 1

	executor := &dummyExecutor{}
	workerQueue := make(WorkerQueue, 1)
	schedulerWorkPool := rp.NewPooler(rp.WithSize[schedulerWork](1))
	requestPool := rp.NewPooler(rp.WithSize[request](1))

	worker := NewWorker(
		workerId,
		executor,
		workerQueue,
		schedulerWorkPool,
		requestPool,
		respPoolSize,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// start the worker
	go func() {
		worker.Start(ctx)
	}()

	// create a scheduler work
	work := &schedulerWork{
		next: func(ctx context.Context, resp core.IResponseReader) {
		},
		request: &request{
			method: "GET",
			header: make(http.Header),
		},
	}
	// execute a task
	worker.execute(ctx, work)
	cancel()
}
