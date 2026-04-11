package core

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/tech-engine/goscrapy/internal/fsmap"
	"golang.org/x/net/html"
)

type IEngine[OUT any] interface {
	Start(context.Context) error
	NewRequest(context.Context) IRequestRW
	Schedule(IRequestReader, ResponseCallback)
	Yield(IOutput[OUT])
	WithName(string)
}

type IRequestReader interface {
	ReadContext() context.Context
	ReadUrl() *url.URL
	ReadHeader() http.Header
	ReadMethod() string
	ReadBody() io.ReadCloser
	ReadMeta() *fsmap.FixedSizeMap[string, any]
	ReadCookieJar() string
}

type IRequestWriter interface {
	Context(context.Context) IRequestWriter
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
	Detach() IResponseReader
	ISelector
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
type ISelectorGetter interface {
	Get() *html.Node
	GetAll() []*html.Node
	Text(...string) []string
	Attr(string) []string
}

type ISelector interface {
	Css(string) ISelector
	Xpath(string) ISelector
	ISelectorGetter
}

// moved from engine package
type IResponseWriter interface {
	WriteHeader(http.Header)
	WriteBody(io.ReadCloser)
	WriteStatusCode(int)
	WriteCookies([]*http.Cookie)
	WriteRequest(*http.Request)
	WriteMeta(*fsmap.FixedSizeMap[string, any])
}

// moved from engine package
type IResponse interface {
	IResponseReader
	IResponseWriter
}
