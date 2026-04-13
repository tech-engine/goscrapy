package httpnative

import (
	"fmt"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/executor"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// HTTPAdapter implements Executor's ExecAdapter interface
type HTTPAdapter struct {
	client *http.Client
	logger core.ILogger
}

func NewHTTPClientAdapter(client *http.Client) *HTTPAdapter {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPAdapter{
		client: client,
		logger: logger.EnsureLogger(nil).WithName("HTTPAdapter"),
	}
}

func (r *HTTPAdapter) WithClient(client *http.Client) {
	r.client = client
}

func (r *HTTPAdapter) WithLogger(loggerIn core.ILogger) executor.IExecutorAdapter {
	loggerIn = logger.EnsureLogger(loggerIn)
	r.logger = loggerIn.WithName("HTTPAdapter")
	return r
}

func (r *HTTPAdapter) Do(res core.IResponseWriter, req *http.Request) error {
	r.logger.Debugf("📡 Sending %s request: %s", req.Method, req.URL.String())
	source, err := r.client.Do(req)

	if err != nil {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("HTTP error: %v", err)
	}

	r.logger.Debugf("✅ Response received: %d %s", source.StatusCode, req.URL.String())
	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
