package engine

import "errors"

var (
	ErrAlreadyStarted         = errors.New("engine already started")
	ErrSchedulerMissing       = errors.New("scheduler is required")
	ErrPipelineManagerMissing = errors.New("pipeline manager is required")
	ErrWorkerPoolMissing      = errors.New("worker pool is required")
	ErrNoCallbacksFound = errors.New("engine: no valid callback methods found in spider")

)
