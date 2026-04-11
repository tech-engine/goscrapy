package httpnative

import (
	"fmt"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// HTTPAdapter implements Executor's ExecAdapter interface
type HTTPAdapter struct {
	client *http.Client
	logger core.ILogger
}

func NewHTTPClientAdapter(client *http.Client, poolSize uint64) *HTTPAdapter {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPAdapter{
		client: client,
		logger: logger.GetLogger(), // default to global logger
	}
}

func (r *HTTPAdapter) WithClient(client *http.Client) {
	r.client = client
}

func (r *HTTPAdapter) WithLogger(logger core.ILogger) {
	r.logger = logger
}

func (r *HTTPAdapter) Do(res core.IResponseWriter, req *http.Request) error {
	r.logger.Debugf("📡 Sending %s request: %s", req.Method, req.URL.String())
	source, err := r.client.Do(req)

	if err != nil {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return fmt.Errorf("Do: request aborted %w", ctxErr)
		}
		r.logger.Errorf("❌ Request failed: %v", err)
		return fmt.Errorf("Do: error dispatching request %w", err)
	}

	r.logger.Debugf("📩 Response received: %d %s", source.StatusCode, req.URL.String())
	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
