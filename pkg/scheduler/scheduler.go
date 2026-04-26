package scheduler

import (
	"context"
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
	stopping  atomic.Bool
	logger    core.ILogger
}

func New(config *Config) (engine.IScheduler, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(nil).WithName("Scheduler")
	}

	if config.WorkQueueSize == 0 {
		config.WorkQueueSize = 1000
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

	return s.taskQueue.Push(context.Background(), &QueuedTask{
		Request:      req,
		CallbackName: cbName,
	})
}

func (s *scheduler) NextRequest(ctx context.Context) (*core.Request, string, engine.TaskHandle, error) {
	qTask, handle, err := s.taskQueue.Pull(ctx)

	if err != nil {
		return nil, "", nil, err
	}

	if qTask == nil {
		return nil, "", nil, nil
	}

	return qTask.Request, qTask.CallbackName, handle, nil
}

// currently passing is a background context,
// but will see later if we need need another context
func (s *scheduler) Ack(handle engine.TaskHandle) error {
	return s.taskQueue.Ack(context.Background(), handle)
}

func (s *scheduler) Nack(handle engine.TaskHandle) error {
	return s.taskQueue.Nack(context.Background(), handle)
}
