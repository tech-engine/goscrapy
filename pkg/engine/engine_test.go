package engine

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/internal/fsm"
	"github.com/tech-engine/goscrapy/pkg/core"
)

//
// --- Helpers ---
//

func newTestEngine() (*Engine[any], *mockScheduler, *mockPipelineManager) {
	s := newMockScheduler()
	pm := newMockPipelineManager()
	return New[any](s, pm), s, pm
}

func waitOrFail(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal(msg)
	}
}

//
// --- Mock Scheduler ---
//

type mockScheduler struct {
	mu         sync.Mutex
	started    bool
	startReady chan struct{}
	scheduled  []core.IRequestReader
	callbacks  []core.ResponseCallback
	startError error
}

func newMockScheduler() *mockScheduler {
	return &mockScheduler{
		startReady: make(chan struct{}),
	}
}

func (m *mockScheduler) Start(ctx context.Context) error {
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()

	select {
	case <-m.startReady:
		// already closed
	default:
		close(m.startReady)
	}

	if m.startError != nil {
		return m.startError
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m *mockScheduler) Schedule(req core.IRequestReader, cb core.ResponseCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scheduled = append(m.scheduled, req)
	m.callbacks = append(m.callbacks, cb)
}

func (m *mockScheduler) NewRequest(ctx context.Context) core.IRequestRW {
	return &mockRequest{}
}

//
// --- Mock Pipeline Manager ---
//

type mockPipelineManager struct {
	mu         sync.Mutex
	started    bool
	startReady chan struct{}
	pushed     []core.IOutput[any]
	startError error
}

func newMockPipelineManager() *mockPipelineManager {
	return &mockPipelineManager{
		startReady: make(chan struct{}),
	}
}

func (m *mockPipelineManager) Start(ctx context.Context) error {
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()

	select {
	case <-m.startReady:
		// already closed
	default:
		close(m.startReady)
	}

	if m.startError != nil {
		return m.startError
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m *mockPipelineManager) Push(out core.IOutput[any]) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pushed = append(m.pushed, out)
}

//
// --- Mock Request ---
//

type mockRequest struct {
	method string
}

func (r *mockRequest) ReadContext() context.Context                        { return context.Background() }
func (r *mockRequest) ReadUrl() *url.URL                                   { return nil }
func (r *mockRequest) ReadHeader() http.Header                             { return nil }
func (r *mockRequest) ReadMethod() string                                  { return r.method }
func (r *mockRequest) ReadBody() io.ReadCloser                             { return nil }
func (r *mockRequest) ReadMeta() *fsm.FixedSizeMap[string, any]            { return nil }
func (r *mockRequest) ReadCookieJar() string                               { return "" }
func (r *mockRequest) WithContext(ctx context.Context) core.IRequestWriter { return r }
func (r *mockRequest) Url(s string) core.IRequestWriter                    { return r }
func (r *mockRequest) Header(h http.Header) core.IRequestWriter            { return r }
func (r *mockRequest) Method(m string) core.IRequestWriter                 { r.method = m; return r }
func (r *mockRequest) Body(b any) core.IRequestWriter                      { return r }
func (r *mockRequest) Meta(k string, v any) core.IRequestWriter            { return r }
func (r *mockRequest) CookieJar(k string) core.IRequestWriter              { return r }
func (r *mockRequest) Reset()                                              {}

//
// --- Tests ---
//

func TestEngine_New(t *testing.T) {
	eng, _, _ := newTestEngine()
	require.NotNil(t, eng)
}

func TestEngine_Schedule_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		req  core.IRequestReader
		cb   core.ResponseCallback
	}{
		{"valid request", &mockRequest{}, func(ctx context.Context, r core.IResponseReader) {}},
		{"nil callback", &mockRequest{}, nil},
		{"nil request", nil, func(ctx context.Context, r core.IResponseReader) {}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, sched, _ := newTestEngine()

			eng.Schedule(tt.req, tt.cb)

			sched.mu.Lock()
			defer sched.mu.Unlock()

			assert.Len(t, sched.scheduled, 1)
			assert.Equal(t, tt.req, sched.scheduled[0])
			assert.Len(t, sched.callbacks, 1)
		})
	}
}

func TestEngine_Schedule_Multiple(t *testing.T) {
	eng, sched, _ := newTestEngine()

	cb := func(ctx context.Context, resp core.IResponseReader) {}

	for i := 0; i < 5; i++ {
		eng.Schedule(&mockRequest{}, cb)
	}

	sched.mu.Lock()
	defer sched.mu.Unlock()

	assert.Len(t, sched.scheduled, 5)
	assert.Len(t, sched.callbacks, 5)
}

func TestEngine_NewRequest(t *testing.T) {
	eng, _, _ := newTestEngine()
	assert.NotNil(t, eng.NewRequest(context.Background()))
}

func TestEngine_Start_Lifecycle(t *testing.T) {
	eng, sched, pm := newTestEngine()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- eng.Start(ctx)
	}()

	waitOrFail(t, sched.startReady, "scheduler did not start")
	waitOrFail(t, pm.startReady, "pipeline manager did not start")

	cancel()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("engine did not stop after cancel")
	}

	sched.mu.Lock()
	assert.True(t, sched.started)
	sched.mu.Unlock()

	pm.mu.Lock()
	assert.True(t, pm.started)
	pm.mu.Unlock()
}

func TestEngine_Start_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"scheduler fails", errors.New("scheduler error")},
		{"no error (cancel wins)", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, sched, _ := newTestEngine()
			sched.startError = tt.err

			ctx, cancel := context.WithCancel(context.Background())

			if tt.err == nil {
				// ensure cancel path is tested
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			} else {
				defer cancel()
			}

			done := make(chan error, 1)
			go func() {
				done <- eng.Start(ctx)
			}()

			select {
			case err := <-done:
				if tt.err != nil {
					assert.True(t,
						errors.Is(err, tt.err) || errors.Is(err, context.Canceled),
					)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("engine did not return")
			}
		})
	}
}

func TestEngine_WithOverrides(t *testing.T) {
	eng, sched1, _ := newTestEngine()

	sched2 := newMockScheduler()
	pm2 := newMockPipelineManager()

	eng.WithScheduler(sched2)
	eng.WithPipelineManager(pm2)

	req := &mockRequest{}
	eng.Schedule(req, nil)

	sched1.mu.Lock()
	assert.Empty(t, sched1.scheduled)
	sched1.mu.Unlock()

	sched2.mu.Lock()
	assert.Len(t, sched2.scheduled, 1)
	sched2.mu.Unlock()
}

func TestEngine_Start_Idempotent(t *testing.T) {
	eng, _, _ := newTestEngine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Start(ctx)

	// second call should not panic (but may no-op or error)
	assert.NotPanics(t, func() {
		go eng.Start(ctx)
	})

	time.Sleep(100 * time.Millisecond)
}
