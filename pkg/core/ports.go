package core

import (
	"context"
	"io"
	"net/http"

	"github.com/tech-engine/goscrapy/internal/fsmap"
	"golang.org/x/net/html"
)

type IActivityTracker interface {
	Inc()
	Dec()
}

type IRequestPool interface {
	Acquire(context.Context) *Request
	Release(*Request)
}

type IEngine[OUT any] interface {
	Start(context.Context) error
	Schedule(*Request, ResponseCallback)
	Yield(IOutput[OUT])
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

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

type ILogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	// must return a logger pointing to the same writer as that of parent
	WithName(name string) ILogger
}

// IConfigurableLogger is the framework-level interface that allows output redirection.
type IConfigurableLogger interface {
	ILogger
	WithWriter(io.Writer) ILogger
}
