package scheduler

import (
	"context"
	"sync"
)

// Worker will handle the execution of a Work unit
type Worker struct {
	ID          uint16
	executor    IExecutor
	workerQueue WorkerQueue
	workQueue   WorkQueue
	quit        chan struct{}
}

func NewWorker(id uint16, executor IExecutor, workerQueue WorkerQueue) *Worker {

	return &Worker{
		ID:          id,
		workerQueue: workerQueue,
		executor:    executor,
		workQueue:   make(WorkQueue),
	}
}

// Handles listen for any incoming work in workQueue
func (w *Worker) Start(ctx context.Context) error {
	var err error

	if err = ctx.Err(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	// we wait for all worker jobs to be completed finished/fail afer context cancellation
	defer wg.Wait()

	for {

		// make this worker available again
		w.workerQueue <- w.workQueue

		select {
		case work := <-w.workQueue:

			if err = ctx.Err(); err != nil {
				return err
			}

			wg.Add(1)

			// we don't want the workers to crash, so we ignore the error from execute
			_ = w.execute(ctx, work)
			wg.Done()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Handles executing a scheduler work and calling the next callback of with the result as response
func (w *Worker) execute(ctx context.Context, work *schedulerWork) error {

	res := responsePool.Acquire()

	if res == nil {
		res = &response{}
	}

	// we do some cleanup here on the response object
	defer func() {
		res.body.Close()
		res.Reset()
		responsePool.Release(res)
	}()

	if err := w.executor.Execute(work.request, res); err != nil {
		resetAndRelease(work)
		return err
	}

	next := (*work).next
	pCtx := work.request.ReadContext()

	resetAndRelease(work)

	// next==nil means this is the last callback of the spider
	if next == nil {
		return nil
	}
	
	// call to callback must me blocking so that the callback can read from the response
	// before the response is resetted and returned to pool

	if pCtx == nil {
		pCtx = context.Background()
	}

	next(context.WithValue(pCtx, "WORKER_ID", w.ID), res)
	return nil
}

func resetAndRelease(work *schedulerWork) {
	// release *request to pool
	req, ok := work.request.(*request)

	if !ok {
		return
	}

	req.Reset()

	requestPool.Release(req)

	// release *schedulerWork to pool
	work.Reset()

	schedulerWorkPool.Release(work)
}
