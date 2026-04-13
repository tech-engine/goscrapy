package httpnative

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/executor"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// httpAdapter implements Executor's ExecAdapter interface
type httpAdapter struct {
	client *http.Client
	logger core.ILogger
}

type adapterOpts struct {
	client *http.Client
	logger core.ILogger
}

func defaultOpts() adapterOpts {
	return adapterOpts{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
		logger: logger.NewLogger(),
	}
}

func WithClient(cli *http.Client) types.OptFunc[adapterOpts] {
	return func(o *adapterOpts) {
		o.client = cli
	}
}

func NewAdapter(opts ...types.OptFunc[adapterOpts]) *httpAdapter {
	options := defaultOpts()

	for _, opt := range opts {
		opt(&options)
	}

	return &httpAdapter{
		client: options.client,
		logger: logger.EnsureLogger(options.logger).WithName("HTTPAdapter"),
	}
}

func (a *httpAdapter) WithLogger(loggerIn core.ILogger) executor.IExecutorAdapter {
	loggerIn = logger.EnsureLogger(loggerIn)
	a.logger = loggerIn.WithName("HTTPAdapter")
	return a
}

func (a *httpAdapter) Do(res core.IResponseWriter, req *http.Request) error {
	a.logger.Debugf("📡 Sending %s request: %s", req.Method, req.URL.String())
	source, err := a.client.Do(req)

	if err != nil {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("HTTP error: %v", err)
	}

	a.logger.Debugf("✅ Response received: %d %s", source.StatusCode, req.URL.String())
	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
