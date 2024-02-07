package engine

import (
	"context"
	"io"
	"net/http"

	"github.com/tech-engine/goscrapy/internal/fsm"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type IPipelineManager[OUT any] interface {
	Start(context.Context) error
	Push(core.IOutput[OUT])
}

type Resetter interface {
	Reset()
}

type IResponseWriter interface {
	WriteHeader(http.Header)
	WriteBody(io.ReadCloser)
	WriteStatusCode(int)
	WriteCookies([]*http.Cookie)
	WriteRequest(*http.Request)
	WriteMeta(*fsm.FixedSizeMap[string, any])
}

type IResponse interface {
	core.IResponseReader
	IResponseWriter
}

type IScheduler interface {
	Start(context.Context) error
	Schedule(core.IRequestReader, core.ResponseCallback)
	NewRequest() core.IRequestRW
}
