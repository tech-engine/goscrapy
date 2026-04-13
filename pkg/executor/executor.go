package executor

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

type Executor struct {
	adapter IExecutorAdapter
	logger  core.ILogger
}

func New(adapter IExecutorAdapter) *Executor {
	return &Executor{
		adapter: adapter,
		logger:  logger.EnsureLogger(nil).WithName("Executor"),
	}
}

func (e *Executor) Execute(req core.IRequestReader, res core.IResponseWriter) error {

	ctx := req.ReadContext()
	if ctx == nil {
		ctx = context.Background()
	}

	if req.ReadCookieJar() != "" {
		ctx = core.InjectCtxValue(ctx, "GOSCookieJarKey", req.ReadCookieJar())
	}

	method := "GET"
	if req.ReadMethod() != "" {
		method = req.ReadMethod()
	}

	request, err := http.NewRequestWithContext(ctx, method, req.ReadUrl().String(), req.ReadBody())
	if err != nil {
		return err
	}

	request.URL = req.ReadUrl()
	request.Header = req.ReadHeader()

	return e.adapter.Do(res, request)
}

func (e *Executor) WithAdapter(adapter IExecutorAdapter) {
	e.adapter = adapter
}

func (e *Executor) WithLogger(loggerIn core.ILogger) scheduler.IExecutor {
	loggerIn = logger.EnsureLogger(loggerIn)
	e.logger = loggerIn.WithName("Executor")
	e.adapter.WithLogger(loggerIn)
	return e
}
