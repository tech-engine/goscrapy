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
	tracker          core.IActivityTracker
	logger           core.ILogger
	pool             *workerPool
}

func NewWorker(id uint16, executor IExecutor, workerTaskBuffer <-chan *workTask, results chan<- engine.IResult, responsePool *sync.Pool, taskPool *sync.Pool, pool *workerPool) *Worker {
	return &Worker{
		ID:               id,
		executor:         executor,
		workerTaskBuffer: workerTaskBuffer,
		results:          results,
		responsePool:     responsePool,
		workerTaskPool:   taskPool,
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

	// context will be cancelled in ReleaseResult after body is drained
	execCtx, cancel := context.WithCancel(ctx)

	// use request deadline if provided
	if task.req.Ctx != nil {
		if d, ok := task.req.Ctx.Deadline(); ok {
			execCtx, _ = context.WithDeadline(execCtx, d)
		}
	}

	// set our exec context
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

	return &result{
		request:      task.req,
		response:     resp,
		callbackName: task.callbackName,
		taskHandle:   task.taskHandle,
		err:          err,
		cancel:       cancel,
	}
}
