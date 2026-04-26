package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type mockScheduler struct {
	started bool
}

func (m *mockScheduler) Start(ctx context.Context) error {
	m.started = true
	<-ctx.Done()
	return ctx.Err()
}
func (m *mockScheduler) Schedule(req *core.Request, cbName string) error { return nil }
func (m *mockScheduler) NextRequest(ctx context.Context) (*core.Request, string, core.TaskHandle, error) {
	<-ctx.Done()
	return nil, "", nil, ctx.Err()
}
func (m *mockScheduler) Ack(handle core.TaskHandle) error  { return nil }
func (m *mockScheduler) Nack(handle core.TaskHandle) error { return nil }

type mockWorkerPool struct {
	started bool
}

func (m *mockWorkerPool) Start(ctx context.Context) error {
	m.started = true
	<-ctx.Done()
	return ctx.Err()
}
func (m *mockWorkerPool) ReleaseResult(res IResult) {}
func (m *mockWorkerPool) Submit(req *core.Request, cbName string, handle core.TaskHandle) error {
	return nil
}
func (m *mockWorkerPool) Results() <-chan IResult {
	return make(chan IResult)
}

type mockPipelineManager struct {
	started bool
}

func (m *mockPipelineManager) Start(ctx context.Context) error {
	m.started = true
	<-ctx.Done()
	return ctx.Err()
}
func (m *mockPipelineManager) Push(out core.IOutput[any])      {}
func (m *mockPipelineManager) Add(pipelines ...IPipeline[any]) {}
func (m *mockPipelineManager) Stop()                           {}

func TestEngine_Lifecycle(t *testing.T) {
	s := &mockScheduler{}
	wp := &mockWorkerPool{}
	pm := &mockPipelineManager{}

	eng, err := New(&Config[any]{
		Scheduler:       s,
		WorkerPool:      wp,
		PipelineManager: pm,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = eng.Start(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	assert.True(t, s.started)
	assert.True(t, wp.started)
	assert.True(t, pm.started)
}

type mySpider struct{}

func (s *mySpider) Parse(ctx context.Context, r core.IResponseReader) {}

type badSpider struct{}

func TestEngine_RegisterSpider(t *testing.T) {
	eng, err := New(&Config[any]{
		Scheduler:       &mockScheduler{},
		WorkerPool:      &mockWorkerPool{},
		PipelineManager: &mockPipelineManager{},
	})
	require.NoError(t, err)

	err = eng.RegisterSpider(&mySpider{})
	assert.NoError(t, err)

	// Test invalid spider (no parse method)
	err = eng.RegisterSpider(&badSpider{})
	assert.Error(t, err)
}
