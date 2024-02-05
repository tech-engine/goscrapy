package executor

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type IExecutorRequestWriter interface {
	WithContext(context.Context)
	Header(http.Header)
	Body(io.ReadCloser)
}

type IExecutorRequestMaker interface {
	Get(engine.IResponseWriter, *url.URL) error
	Post(engine.IResponseWriter, *url.URL) error
	Patch(engine.IResponseWriter, *url.URL) error
	Put(engine.IResponseWriter, *url.URL) error
	Delete(engine.IResponseWriter, *url.URL) error
}

type IExecutorAdapter interface {
	IExecutorRequestWriter
	IExecutorRequestMaker
}
