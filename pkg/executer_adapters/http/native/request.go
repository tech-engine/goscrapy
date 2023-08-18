package nativeadapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// NativeHTTPRequestAdapter implements executer's Requester interface

type NativeHTTPRequestAdapter struct {
	client *http.Client
	req    *http.Request
}

func (r NativeHTTPRequestAdapter) SetContext(ctx context.Context) executorhttp.Requester {
	r.req = r.req.WithContext(ctx)
	return r
}

func (r NativeHTTPRequestAdapter) SetHeaders(headers map[string]string) executorhttp.Requester {
	for key, value := range headers {
		r.req.Header.Add(key, value)
	}
	return r
}

func (r NativeHTTPRequestAdapter) SetBody(body io.ReadCloser) executorhttp.Requester {
	r.req.Body = body
	return r
}

func (r NativeHTTPRequestAdapter) Get(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodGet
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Get: error dispatching request %w", err)
	}

	return NativeHTTPRequestAdapterResponse(target, source, err)
}

func (r NativeHTTPRequestAdapter) Post(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPost
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Post: error dispatching request %w", err)
	}

	return NativeHTTPRequestAdapterResponse(target, source, err)
}

func (r NativeHTTPRequestAdapter) Put(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPut
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Put: error dispatching request %w", err)
	}

	return NativeHTTPRequestAdapterResponse(target, source, err)
}

func (r NativeHTTPRequestAdapter) Patch(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPatch
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Patch: error dispatching request %w", err)
	}

	return NativeHTTPRequestAdapterResponse(target, source, err)
}

func (r NativeHTTPRequestAdapter) Delete(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodDelete
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Delete: error dispatching request %w", err)
	}

	return NativeHTTPRequestAdapterResponse(target, source, err)
}
