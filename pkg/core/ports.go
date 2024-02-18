package core

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/tech-engine/goscrapy/internal/fsm"
)

type IEngine[OUT any] interface {
	Start(context.Context) error
	NewRequest() IRequestRW
	Schedule(IRequestReader, ResponseCallback)
	Yield(IOutput[OUT])
}

type IRequestReader interface {
	ReadContext() context.Context
	ReadUrl() *url.URL
	ReadHeader() http.Header
	ReadMethod() string
	ReadBody() io.ReadCloser
	ReadMeta() *fsm.FixedSizeMap[string, any]
	ReadCookieJar() string
}

type IRequestWriter interface {
	WithContext(context.Context) IRequestWriter
	Url(string) IRequestWriter
	Header(http.Header) IRequestWriter
	Method(string) IRequestWriter
	Body(any) IRequestWriter
	Meta(string, any) IRequestWriter
	CookieJar(string) IRequestWriter
}

type IRequestRW interface {
	IRequestReader
	IRequestWriter
	Reset()
}

type IResponseReader interface {
	Header() http.Header
	Body() io.ReadCloser
	Bytes() []byte
	StatusCode() int
	Cookies() []*http.Cookie
	Request() *http.Request
	Meta(string) (any, bool)
}

type IJob interface {
	Id() string
}

type IOutput[OUT any] interface {
	Record() OUT
	RecordKeys() []string
	RecordFlat() []any
	Job() IJob
}

type ResponseCallback func(context.Context, IResponseReader)
