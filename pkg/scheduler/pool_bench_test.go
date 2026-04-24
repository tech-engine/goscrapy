// Note: generated benchmark test
package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/tech-engine/goscrapy/internal/request"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type benchExecutor struct {
	client *http.Client
}

func (e *benchExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	request, _ := http.NewRequestWithContext(context.Background(), "GET", req.URL.String(), nil)
	resp, err := e.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	res.WriteStatusCode(resp.StatusCode)
	return nil
}

func (e *benchExecutor) WithLogger(logger core.ILogger) IExecutor { return e }

// BenchmarkSchedulerPooling measures allocation overhead of the full scheduler pipeline.
func BenchmarkSchedulerPooling(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		sched := New(executor, request.NewPool(),
			WithWorkers(4),
			WithReqResPoolSize(128),
			WithWorkQueueSize(128),
		)

		go sched.Start(ctx)

		req := request.NewPool().Acquire(ctx)
		req.URL, _ = url.Parse(ts.URL + "/")

		done := make(chan struct{})
		sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
			close(done)
		})

		<-done
		cancel()
	}
}
