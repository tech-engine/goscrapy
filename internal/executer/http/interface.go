package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type RequestReader interface {
	Url() *url.URL
	Headers() map[string]string
	Method() string
	Body() io.ReadCloser
	MetaData() map[string]any
}

// Note: any http request to be processed by executer must implement RequestReaderReseter interface
type RequestReaderReseter interface {
	RequestReader
	Reset()
}

// interface for executor client adapter
type Client interface {
	Request() Requester
}

type Requester interface {
	SetContext(context.Context) Requester
	SetHeaders(map[string]string) Requester
	SetBody(io.ReadCloser) Requester
	Get(ResponseWriter, *url.URL) error
	Post(ResponseWriter, *url.URL) error
	Patch(ResponseWriter, *url.URL) error
	Put(ResponseWriter, *url.URL) error
	Delete(ResponseWriter, *url.URL) error
}

type ResponseWriter interface {
	SetStatusCode(int) ResponseWriter
	SetHeaders(http.Header) ResponseWriter
	SetBody(io.ReadCloser) ResponseWriter
	SetCookies([]*http.Cookie) ResponseWriter
}
