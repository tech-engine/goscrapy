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

func (e *Executor) Execute(req *core.Request, res core.IResponseWriter) error {

	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if req.CookieJarKey != "" {
		ctx = core.InjectCtxValue(ctx, "GOSCookieJarKey", req.CookieJarKey)
	}

	method := "GET"
	if req.Method != "" {
		method = req.Method
	}

	request, err := http.NewRequestWithContext(ctx, method, req.URL.String(), req.Body)
	if err != nil {
		return err
	}

	request.Header = req.Header

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
