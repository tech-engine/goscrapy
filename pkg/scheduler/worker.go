package scheduler

import (
	"context"
	"io"
	"sync"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
)

// Worker will handle the execution of a Work unit
type Worker struct {
	ID                uint16
	executor          IExecutor
	workerQueue       WorkerQueue
	workQueue         WorkQueue
	schedulerWorkPool *rp.Pooler[schedulerWork]
	responsePool      *rp.Pooler[response]
	requestPool       *rp.Pooler[request]
}

func NewWorker(id uint16, executor IExecutor, workerQueue WorkerQueue, schedulerWorkPool *rp.Pooler[schedulerWork], requestPool *rp.Pooler[request], respPoolSize uint64) *Worker {

	return &Worker{
		ID:                id,
		workerQueue:       workerQueue,
		executor:          executor,
		workQueue:         make(WorkQueue),
		schedulerWorkPool: schedulerWorkPool,
		requestPool:       requestPool,
		responsePool:      rp.NewPooler(rp.WithSize[response](respPoolSize)),
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

	res := w.responsePool.Acquire()

	if res == nil {
		res = &response{}
	}

	// we do some cleanup here on the response object
	defer func() {
		w.resetAndRelease(work)
		// discard unread body
		io.Copy(io.Discard, res.body)
		res.body.Close()
		res.Reset()
		w.responsePool.Release(res)
	}()

	if err := w.executor.Execute(work.request, res); err != nil {
		// resetAndRelease(work)
		return err
	}

	next := (*work).next
	pCtx := work.request.ReadContext()

	// next==nil means this is the last callback of the spider
	if next == nil {
		return nil
	}

	// call to callback must me blocking so that the callback can read from the response
	// before the response is resetted and returned to pool
	if pCtx == nil {
		pCtx = context.Background()
	}

	// we copy meta from our request to our response to be accessible to the spider
	res.WriteMeta(work.request.ReadMeta())

	next(context.WithValue(pCtx, "WORKER_ID", w.ID), res)
	return nil
}

func (w *Worker) resetAndRelease(work *schedulerWork) {
	// release *request to pool
	req, ok := work.request.(*request)

	if !ok {
		return
	}

	req.Reset()

	w.requestPool.Release(req)

	// release *schedulerWork to pool
	work.Reset()

	w.schedulerWorkPool.Release(work)
}
