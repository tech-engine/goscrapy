package nativeadapter

import (
	"net/http"

	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// NativeHTTPClientAdapter implements executer's Client interface

type NativeHTTPClientAdapter struct {
	client *http.Client
}

func NewHTTPClientAdapter() *NativeHTTPClientAdapter {
	return &NativeHTTPClientAdapter{
		client: &http.Client{},
	}
}

func (r *NativeHTTPClientAdapter) Request() executorhttp.Requester {
	return NativeHTTPRequestAdapter{
		client: r.client,
		req:    &http.Request{},
	}
}
