package scheduler

import "errors"

var (
	ErrSchedulerStopping    = errors.New("scheduler is stopping")
	ErrSchedulerQueueFull   = errors.New("scheduler work queue full")
	ErrSchedulerQueueClosed = errors.New("scheduler work queue closed")
	ErrTaskQueueClosed      = errors.New("task queue closed")
	ErrTaskQueueFull        = errors.New("task queue full")
	ErrFailedGetTask        = errors.New("failed to get task")
)
