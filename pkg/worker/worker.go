package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

type workTask struct {
	req          *core.Request
	callbackName string
	taskHandle   core.TaskHandle
}

func (wt *workTask) Reset() {
	wt.req = nil
	wt.callbackName = ""
	wt.taskHandle = nil
}

type Worker struct {
	ID               uint16
	executor         IExecutor
	workerTaskBuffer <-chan *workTask
	results          chan<- engine.IResult
	workerTaskPool   *sync.Pool
	responsePool     *sync.Pool
	resultPool       *sync.Pool
	tracker          core.IActivityTracker
	logger           core.ILogger
	pool             *workerPool
}

func NewWorker(id uint16, executor IExecutor, workerTaskBuffer <-chan *workTask, results chan<- engine.IResult, responsePool *sync.Pool, taskPool *sync.Pool, resultPool *sync.Pool, pool *workerPool) *Worker {
	return &Worker{
		ID:               id,
		executor:         executor,
		workerTaskBuffer: workerTaskBuffer,
		results:          results,
		responsePool:     responsePool,
		workerTaskPool:   taskPool,
		resultPool:       resultPool,
		pool:             pool,
		logger:           pool.logger.WithName(fmt.Sprintf("Worker-%d", id)),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case task, ok := <-w.workerTaskBuffer:
			if !ok || task == nil || task.req == nil {
				return nil
			}

			wg.Add(1)

			res := w.execute(ctx, task)

			// return task to pool
			task.Reset()
			w.workerTaskPool.Put(task)

			if w.results != nil {
				select {
				case w.results <- res:
				case <-ctx.Done():
					wg.Done()
					return ctx.Err()
				}
			}

			wg.Done()
		}
	}
}

func (w *Worker) execute(ctx context.Context, task *workTask) engine.IResult {
	resp := w.responsePool.Get().(*response)

	// merge contexts to handle timeouts and graceful shutdown
	execCtx, cleanup := mergeContexts(ctx, task.req.Ctx)

	task.req.Ctx = execCtx

	if w.tracker != nil {
		w.tracker.Inc()
		defer w.tracker.Dec()
	}

	start := time.Now()
	err := w.executor.Execute(task.req, resp)

	if w.pool.autoscaler != nil {
		w.pool.autoscaler.OnTaskDone(time.Since(start))
	}

	// we transfer request meta to response so callbacks can access it
	if task.req.Meta_ != nil {
		resp.WriteMeta(task.req.Meta_)
	}

	if w.pool.signals != nil {
		if err != nil {
			w.pool.signals.EmitRequestError(context.Background(), task.req, err)
		} else {
			w.pool.signals.EmitResponseReceived(context.Background(), resp)
		}
	}

	res := w.resultPool.Get().(*result)
	res.request = task.req
	res.response = resp
	res.callbackName = task.callbackName
	res.taskHandle = task.taskHandle
	res.err = err
	res.cancel = cleanup

	return res
}
