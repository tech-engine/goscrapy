package executor

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type Config struct {
	Adapter IExecutorAdapter
	Logger  core.ILogger
}

type Executor struct {
	adapter IExecutorAdapter
	logger  core.ILogger
}

func New(config *Config) (*Executor, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(nil).WithName("Executor")
	}

	if config.Adapter == nil {
		return nil, ErrAdapterRequired
	}

	return &Executor{
		adapter: config.Adapter,
		logger:  config.Logger,
	}, nil
}

func (e *Executor) Execute(req *core.Request, res core.IResponseWriter) error {
	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if req.CookieJarKey != "" {
		ctx = core.InjectCtxValue(ctx, "GOSCookieJarKey", req.CookieJarKey)
	}

	method := "GET"
	if req.Method_ != "" {
		method = req.Method_
	}

	request, err := http.NewRequestWithContext(ctx, method, req.URL.String(), req.Body_)
	if err != nil {
		return err
	}

	request.Header = req.Header_

	return e.adapter.Do(res, request)
}
