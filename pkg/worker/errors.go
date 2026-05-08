package worker

import "errors"

var (
	ErrWorkerPoolFull   = errors.New("worker pool task buffer full")
	ErrExecutorRequired = errors.New("workerpool: executor is required in config")
	ErrPoolClosed       = errors.New("workerpool: pool is closed")
)
