package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/signal"
)

type signalSpider struct {
	opened bool
	idled  bool
	closed bool
	errored bool
}

func (s *signalSpider) Open(ctx context.Context)  { s.opened = true }
func (s *signalSpider) Idle(ctx context.Context)  { s.idled = true }
func (s *signalSpider) Close(ctx context.Context) { s.closed = true }
func (s *signalSpider) Error(ctx context.Context, err error) { s.errored = true }
func (s *signalSpider) Parse(ctx context.Context, r core.IResponseReader) {}

func TestEngine_SignalAutoDiscovery(t *testing.T) {
	eng, err := New(&Config[any]{
		Scheduler:       &mockScheduler{},
		WorkerPool:      &mockWorkerPool{},
		PipelineManager: &mockPipelineManager{},
	})
	require.NoError(t, err)

	spider := &signalSpider{}
	err = eng.RegisterSpider(spider)
	assert.NoError(t, err)

	ctx := context.Background()

	eng.signals.EmitSpiderOpened(ctx)
	assert.True(t, spider.opened, "Open method should have been auto-discovered")

	eng.signals.EmitSpiderIdle(ctx)
	assert.True(t, spider.idled, "Idle method should have been auto-discovered")

	eng.signals.EmitSpiderClosed(ctx)
	assert.True(t, spider.closed, "Close method should have been auto-discovered")

	eng.signals.EmitSpiderError(ctx, errors.New("test"))
	assert.True(t, spider.errored, "Error method should have been auto-discovered")
}

func TestEngine_LifecycleSignals(t *testing.T) {
	bus := signal.New()
	
	startedCalled := false
	stoppedCalled := false
	closedCalled  := false
	
	bus.Connect(signal.EngineStarted, func(ctx context.Context) {
		startedCalled = true
	})
	bus.Connect(signal.EngineStopped, func(ctx context.Context) {
		stoppedCalled = true
	})
	bus.Connect(signal.SpiderClosed, func(ctx context.Context) {
		closedCalled = true
	})

	eng, err := New(&Config[any]{
		Scheduler:       &mockScheduler{},
		WorkerPool:      &mockWorkerPool{},
		PipelineManager: &mockPipelineManager{},
		Signals:         bus,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start engine (will time out due to mocks)
	_ = eng.Start(ctx)

	assert.True(t, startedCalled, "EngineStarted signal should have been emitted")
	assert.True(t, closedCalled, "SpiderClosed signal should have been emitted")
	assert.True(t, stoppedCalled, "EngineStopped signal should have been emitted")
}

func TestEngine_SpiderIdleSignal(t *testing.T) {
	bus := signal.New()
	idleCalled := false
	bus.Connect(signal.SpiderIdle, func(ctx context.Context) {
		idleCalled = true
	})

	eng, err := New(&Config[any]{
		Scheduler:       &mockScheduler{},
		WorkerPool:      &mockWorkerPool{},
		PipelineManager: &mockPipelineManager{},
		Signals:         bus,
	})
	require.NoError(t, err)

	// Simulate engine being in started state
	eng.started.Store(true)

	// Simulate work
	eng.Schedule(&core.Request{}, func(ctx context.Context, r core.IResponseReader) {})
	assert.Equal(t, int64(1), eng.ActiveCount())
	assert.True(t, eng.started.Load())

	// Simulate work completion
	eng.handleResult(context.Background(), &mockResult{})
	
	assert.Equal(t, int64(0), eng.ActiveCount())
	assert.True(t, idleCalled, "SpiderIdle signal should have been emitted when count hit 0")
}

type mockResult struct{}
func (m *mockResult) Request() *core.Request { return nil }
func (m *mockResult) Response() core.IResponseReader { return nil }
func (m *mockResult) CallbackName() string { return "" }
func (m *mockResult) TaskHandle() core.TaskHandle { return nil }
func (m *mockResult) Error() error { return nil }
func (m *mockResult) Release() {}
