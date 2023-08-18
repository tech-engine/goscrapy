package nativeadapter

import (
	"net/http"

	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// NativeHTTPClientAdapter implements executer's Client interface

type HTTPClientAdapter struct {
	client *http.Client
}

func NewHTTPClientAdapter() *HTTPClientAdapter {
	return &HTTPClientAdapter{
		client: &http.Client{},
	}
}

func (r *HTTPClientAdapter) Request() executorhttp.Requester {
	return HTTPRequestAdapter{
		client: r.client,
		req:    &http.Request{},
	}
}
