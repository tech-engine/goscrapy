package executor

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type Executor struct {
	adapter IExecutorAdapter
}

func New(adapter IExecutorAdapter) *Executor {
	return &Executor{
		adapter: adapter,
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

func (e *Executor) WithLogger(logger core.ILogger) {
	e.adapter.WithLogger(logger)
}
