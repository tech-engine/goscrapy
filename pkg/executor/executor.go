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

	if req.ReadContext() != nil {
		e.adapter.WithContext(req.ReadContext())
	}

	e.adapter.Header(req.ReadHeader())
	e.adapter.Body(req.ReadBody())

	switch req.ReadMethod() {
	case "GET":
		return e.adapter.Get(res, req.ReadUrl())
	case "POST":
		return e.adapter.Post(res, req.ReadUrl())
	case "DELETE":
		return e.adapter.Delete(res, req.ReadUrl())
	case "PATCH":
		return e.adapter.Patch(res, req.ReadUrl())
	case "PUT":
		return e.adapter.Put(res, req.ReadUrl())
	default:
		return e.adapter.Get(res, req.ReadUrl())
	}
}

func (e *Executor) WithAdapter(adapter IExecutorAdapter) {
	e.adapter = adapter
}