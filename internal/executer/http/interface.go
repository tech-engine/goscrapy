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

type RequestSetter interface {
	SetContext(context.Context) Requester
	SetHeaders(map[string]string) Requester
	SetBody(io.ReadCloser) Requester
}

type RequestMaker interface {
	Get(ResponseSetter, *url.URL) error
	Post(ResponseSetter, *url.URL) error
	Patch(ResponseSetter, *url.URL) error
	Put(ResponseSetter, *url.URL) error
	Delete(ResponseSetter, *url.URL) error
}

type Requester interface {
	RequestSetter
	RequestMaker
}

type ResponseSetter interface {
	SetStatusCode(int) ResponseSetter
	SetHeaders(http.Header) ResponseSetter
	SetBody(io.ReadCloser) ResponseSetter
	SetCookies([]*http.Cookie) ResponseSetter
}
