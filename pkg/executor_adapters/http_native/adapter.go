package httpnative

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// httpAdapter implements Executor's ExecAdapter interface
type httpAdapter struct {
	client *http.Client
	logger core.ILogger
}

type Config struct {
	Client *http.Client
	Logger core.ILogger
}

func NewAdapter(config *Config) (*httpAdapter, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Client == nil {
		config.Client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		}
	}

	if config.Logger == nil {
		config.Logger = logger.NewLogger()
	}

	return &httpAdapter{
		client: config.Client,
		logger: logger.EnsureLogger(config.Logger).WithName("HTTPAdapter"),
	}, nil
}

func (a *httpAdapter) Do(res core.IResponseWriter, req *http.Request) error {
	a.logger.Debugf("Sending %s request: %s", req.Method, req.URL.String())
	source, err := a.client.Do(req)

	if err != nil {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("HTTP error: %v", err)
	}

	a.logger.Debugf("Response received: %d %s", source.StatusCode, req.URL.String())
	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
