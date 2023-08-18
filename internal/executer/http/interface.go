package http

import (
	"context"
	"io"
	"net/http"
)

type RequestReader interface {
	Url() string
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
	Get(ResponseWriter, string) error
	Post(ResponseWriter, string) error
	Patch(ResponseWriter, string) error
	Put(ResponseWriter, string) error
	Delete(ResponseWriter, string) error
}

type ResponseWriter interface {
	SetStatusCode(int) ResponseWriter
	SetHeaders(http.Header) ResponseWriter
	SetBody(io.ReadCloser) ResponseWriter
}
