package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// MockExecutor is a mock type for IExecutor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	args := m.Called(req, res)
	return args.Error(0)
}

func (m *MockExecutor) WithLogger(l core.ILogger) IExecutor {
	args := m.Called(l)
	return args.Get(0).(IExecutor)
}

func TestScheduler_New(t *testing.T) {
	mockExec := new(MockExecutor)
	s := New(mockExec)

	assert.NotNil(t, s)
	assert.NotNil(t, s.logger)
	assert.Greater(t, s.opts.numWorkers, uint16(0)) // dynamic default based on GOMAXPROCS
}

func TestScheduler_WithLogger(t *testing.T) {
	mockExec := new(MockExecutor)
	s := New(mockExec)

	mockLogger := logger.NewNoopLogger()
	mockExec.On("WithLogger", mock.MatchedBy(func(l core.ILogger) bool { return true })).Return(mockExec)

	s.WithLogger(mockLogger)
	assert.NotNil(t, s.logger)
}

func TestScheduler_Lifecycle(t *testing.T) {
	mockExec := new(MockExecutor)
	s := New(mockExec, WithWorkers(2))

	ctx, cancel := context.WithCancel(context.Background())
	
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start(ctx)
	}()

	// Wait for workers to start and register themselves in workerQueue
	time.Sleep(100 * time.Millisecond)

	// Verify if scheduled jobs are picked up
	req := s.NewRequest(ctx)
	callbackCalled := make(chan struct{})
	
	mockExec.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		close(callbackCalled)
	})

	s.Schedule(req, func(context.Context, core.IResponseReader) {})

	select {
	case <-callbackCalled:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Request was not processed by worker")
	}

	cancel()
	err := <-errCh
	assert.NoError(t, err)
}

func TestScheduler_NewRequest_Pooling(t *testing.T) {
	mockExec := new(MockExecutor)
	s := New(mockExec)

	ctx := context.Background()
	req1 := s.NewRequest(ctx)
	assert.NotNil(t, req1)
	
	// manually release and re-acquire to verify pooling
	s.requestPool.Release(req1.(*request))
	req2 := s.NewRequest(ctx)
	
	assert.Equal(t, req1, req2, "Should have acquired the same request from pool")
}
