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

type HTTPRequestAdapter struct {
	client *http.Client
	req    *http.Request
}

func (r HTTPRequestAdapter) SetContext(ctx context.Context) executorhttp.Requester {
	r.req = r.req.WithContext(ctx)
	return r
}

func (r HTTPRequestAdapter) SetHeaders(headers map[string]string) executorhttp.Requester {
	for key, value := range headers {
		r.req.Header.Add(key, value)
	}
	return r
}

func (r HTTPRequestAdapter) SetBody(body io.ReadCloser) executorhttp.Requester {
	r.req.Body = body
	return r
}

func (r HTTPRequestAdapter) Get(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodGet
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Get: error dispatching request %w", err)
	}

	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Post(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPost
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Post: error dispatching request %w", err)
	}

	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Put(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPut
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Put: error dispatching request %w", err)
	}

	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Patch(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodPatch
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Patch: error dispatching request %w", err)
	}

	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Delete(target executorhttp.ResponseWriter, _url string) error {
	r.req.Method = http.MethodDelete
	r.req.URL, _ = url.Parse(_url)

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Delete: error dispatching request %w", err)
	}

	return HTTPRequestAdapterResponse(target, source, err)
}
