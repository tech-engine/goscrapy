package executor

import (
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

type Executor struct {
	adapter IExecutorAdapter
}

func New(adapter IExecutorAdapter) *Executor {
	return &Executor{
		adapter: adapter,
	}
}

func (e *Executor) Execute(req core.IRequestReader, res engine.IResponseWriter) error {

	request := e.adapter.Acquire()

	if req.ReadContext() != nil {
		request.WithContext(req.ReadContext())
	}

	headers := req.ReadHeader()
	// we inject a header for cookiejar implementation
	headers.Add("X-Goscrapy-Cookie-Jar-Key", req.ReadCookieJar())

	request.URL = req.ReadUrl()
	request.Method = "GET"

	if req.ReadMethod() != "" {
		request.Method = req.ReadMethod()
	}

	request.Header = headers

	request.Body = req.ReadBody()

	return e.adapter.Do(res, request)
}

func (e *Executor) WithAdapter(adapter IExecutorAdapter) {
	e.adapter = adapter
}
