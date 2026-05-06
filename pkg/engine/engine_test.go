package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type engineTestScheduler struct {
	started bool
	stopped atomic.Bool
}

func (s *engineTestScheduler) Start(ctx context.Context) error {
	s.started = true
	<-ctx.Done()
	s.stop(ctx)
	return nil
}
func (s *engineTestScheduler) stop(ctx context.Context) error {
	s.stopped.Store(true)
	return nil
}
func (s *engineTestScheduler) Schedule(req *core.Request, cbName string) error { return nil }
func (s *engineTestScheduler) NextRequest(ctx context.Context) (*core.Request, string, core.TaskHandle, error) {
	<-ctx.Done()
	return nil, "", nil, ctx.Err()
}
func (s *engineTestScheduler) Ack(handle core.TaskHandle) error  { return nil }
func (s *engineTestScheduler) Nack(handle core.TaskHandle) error { return nil }

type engineTestWorkerPool struct {
	started bool
	results chan IResult
}

// mock worker pool
func newEngineTestWorkerPool() *engineTestWorkerPool {
	return &engineTestWorkerPool{
		results: make(chan IResult),
	}
}

func (wp *engineTestWorkerPool) Start(ctx context.Context) error {
	wp.started = true
	<-ctx.Done()
	wp.Stop()
	return nil
}

func (wp *engineTestWorkerPool) Stop() {
	close(wp.results)
}

func (wp *engineTestWorkerPool) ReleaseResult(res IResult) {}
func (wp *engineTestWorkerPool) Submit(req *core.Request, cbName string, handle core.TaskHandle) error {
	return nil
}
func (wp *engineTestWorkerPool) Results() <-chan IResult {
	return wp.results
}

type engineTestPipelineManager struct {
	started bool
}

func (pm *engineTestPipelineManager) Start(ctx context.Context) error {
	pm.started = true
	<-ctx.Done()
	return nil
}
func (pm *engineTestPipelineManager) Push(out core.IOutput[any])      {}
func (pm *engineTestPipelineManager) Add(pipelines ...IPipeline[any]) {}
func (pm *engineTestPipelineManager) Stop()                           {}

func TestEngine_Lifecycle(t *testing.T) {
	s := &engineTestScheduler{}
	wp := newEngineTestWorkerPool()
	pm := &engineTestPipelineManager{}

	eng, err := New(&Config[any]{
		Scheduler:       s,
		WorkerPool:      wp,
		PipelineManager: pm,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = eng.Start(ctx)
	assert.NoError(t, err)

	assert.True(t, s.started)
	assert.True(t, wp.started)
	assert.True(t, pm.started)
}

type mySpider struct{}

func (s *mySpider) Parse(ctx context.Context, r core.IResponseReader) {}

type badSpider struct{}

func TestEngine_RegisterSpider(t *testing.T) {
	eng, err := New(&Config[any]{
		Scheduler:       &engineTestScheduler{},
		WorkerPool:      newEngineTestWorkerPool(),
		PipelineManager: &engineTestPipelineManager{},
	})
	require.NoError(t, err)

	err = eng.RegisterSpider(&mySpider{})
	assert.NoError(t, err)

	// Test invalid spider (no parse method)
	err = eng.RegisterSpider(&badSpider{})
	assert.Error(t, err)
}
func TestEngine_GracefulShutdownWithResults(t *testing.T) {
	s := &engineTestScheduler{}
	wp := newEngineTestWorkerPool()
	pm := &engineTestPipelineManager{}
	cbRegistry := NewCallbackRegistry()

	var callbackCalled atomic.Bool
	cbRegistry.Register("test_cb", func(ctx context.Context, r core.IResponseReader) {
		callbackCalled.Store(true)
	})

	eng, err := New(&Config[any]{
		Scheduler:        s,
		WorkerPool:       wp,
		PipelineManager:  pm,
		CallbackRegistry: cbRegistry,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// start engine
	done := make(chan error, 1)
	go func() {
		done <- eng.Start(ctx)
	}()

	// give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// inject a result into the pool's results channel
	res := &engineTestResult{
		callbackName: "test_cb",
	}
	
	// we use a goroutine to send because the results channel in mock might be unbuffered
	go func() {
		wp.results <- res
		// trigger shutdown after sending result
		// cancel would cancel the context and this would then stop the engine
		// and we would recieve on done channel below.
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// wait for engine to stop
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("Engine did not shut down gracefully")
	}

	// verify the callback was called (meaning the result was drained and handled)
	assert.True(t, callbackCalled.Load(), "Callback should have been executed before engine stopped")
}

type engineTestResult struct {
	callbackName string
}

func (r *engineTestResult) Request() *core.Request { return nil }
func (r *engineTestResult) Response() core.IResponseReader { return nil }
func (r *engineTestResult) CallbackName() string { return r.callbackName }
func (r *engineTestResult) TaskHandle() core.TaskHandle { return nil }
func (r *engineTestResult) Error() error { return nil }
func (r *engineTestResult) Release() {}
