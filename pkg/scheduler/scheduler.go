package scheduler

import (
	"context"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type Config struct {
	WorkQueueSize uint64 // currently mainly used for inMemTaskQueue's buffer size, but can also be used for prefetching hint later
	Logger        core.ILogger
	TaskQueue     ITaskQueue // Default: inMemTaskQueue
}

type QueuedTask struct {
	Request      *core.Request
	CallbackName string
}

type scheduler struct {
	taskQueue ITaskQueue
	logger    core.ILogger
	stopping  atomic.Bool
}

func New(config *Config) (engine.IScheduler, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(nil).WithName("Scheduler")
	}

	if config.WorkQueueSize == 0 {
		if v := os.Getenv("SCHEDULER_WORK_QUEUE_SIZE"); v != "" {
			if q, err := strconv.ParseUint(v, 10, 64); err == nil && q > 0 {
				config.WorkQueueSize = q
			}
		}
		if config.WorkQueueSize == 0 {
			config.WorkQueueSize = SCHEDULER_DEFAULT_WORK_QUEUE_SIZE
		}
	}

	if config.TaskQueue == nil {
		config.TaskQueue = newInMemoryTaskQueue(config.WorkQueueSize)
	}

	s := &scheduler{
		taskQueue: config.TaskQueue,
		logger:    config.Logger,
	}

	return s, nil
}

func (s *scheduler) Start(ctx context.Context) error {
	s.logger.Info("Starting scheduler")

	<-ctx.Done()
	s.stopping.Store(true)

	s.logger.Info("Stopped scheduler")
	return nil
}

func (s *scheduler) Schedule(req *core.Request, cbName string) error {
	if s.stopping.Load() {
		return ErrSchedulerStopping
	}

	task := QueuedTask{
		Request:      req,
		CallbackName: cbName,
	}

	return s.taskQueue.Push(context.Background(), task)
}

func (s *scheduler) NextRequest(ctx context.Context) (*core.Request, string, engine.TaskHandle, error) {
	qTask, handle, err := s.taskQueue.Pull(ctx)

	if err != nil {
		return nil, "", nil, err
	}

	// For pass-by-value, check if the task is essentially "empty"
	if qTask.Request == nil {
		return nil, "", nil, nil
	}

	req := qTask.Request
	cbName := qTask.CallbackName

	return req, cbName, handle, nil
}

// currently passing is a background context,
// but will see later if we need need another context
func (s *scheduler) Ack(handle engine.TaskHandle) error {
	return s.taskQueue.Ack(context.Background(), handle)
}

func (s *scheduler) Nack(handle engine.TaskHandle) error {
	return s.taskQueue.Nack(context.Background(), handle)
}
