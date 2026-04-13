// Note: generated benchmark test
package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type benchExecutor struct {
	client *http.Client
}

func (e *benchExecutor) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	request, _ := http.NewRequestWithContext(context.Background(), "GET", req.ReadUrl().String(), nil)
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
		sched := New(executor,
			WithWorkers(4),
			WithReqResPoolSize(128),
			WithWorkQueueSize(128),
		)

		go sched.Start(ctx)

		req := sched.NewRequest(ctx)
		req.Url(ts.URL + "/")

		done := make(chan struct{})
		sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
			close(done)
		})

		<-done
		cancel()
	}
}
