package httpnative

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

// HTTPAdapter implements Executor's ExecAdapter interface
type HTTPAdapter struct {
	client *http.Client
	req    *http.Request
}

func NewHTTPClientAdapter(client *http.Client) HTTPAdapter {
	if client == nil {
		client = &http.Client{}
	}

	return HTTPAdapter{
		client: client,
		req:    &http.Request{},
	}
}

func (r HTTPAdapter) WithClient(client *http.Client) {
	r.client = client
}

func (r HTTPAdapter) WithContext(ctx context.Context) {
	r.req = r.req.WithContext(ctx)
}

func (r HTTPAdapter) Header(header http.Header) {
	// r.req.Header = http.Header{}
	// for key, value := range headers {
	// 	r.req.Header.Add(key, value)
	// }
	r.req.Header = header
}

func (r HTTPAdapter) Body(body io.ReadCloser) {
	r.req.Body = body
}

func (r HTTPAdapter) Get(res engine.IResponseWriter, url *url.URL) error {
	r.req.Method = http.MethodGet
	r.req.URL = url

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Get: error dispatching request %w", err)
	}

	res.WriteRequest(r.req)
	return HTTPRequestAdapterResponse(res, source, err)
}

func (r HTTPAdapter) Post(res engine.IResponseWriter, url *url.URL) error {
	r.req.Method = http.MethodPost
	r.req.URL = url

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Post: error dispatching request %w", err)
	}

	res.WriteRequest(r.req)
	return HTTPRequestAdapterResponse(res, source, err)
}

func (r HTTPAdapter) Put(res engine.IResponseWriter, url *url.URL) error {
	r.req.Method = http.MethodPut
	r.req.URL = url

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Put: error dispatching request %w", err)
	}

	res.WriteRequest(r.req)
	return HTTPRequestAdapterResponse(res, source, err)
}

func (r HTTPAdapter) Patch(res engine.IResponseWriter, url *url.URL) error {
	r.req.Method = http.MethodPatch
	r.req.URL = url

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Patch: error dispatching request %w", err)
	}

	res.WriteRequest(r.req)
	return HTTPRequestAdapterResponse(res, source, err)
}

func (r HTTPAdapter) Delete(res engine.IResponseWriter, url *url.URL) error {
	r.req.Method = http.MethodDelete
	r.req.URL = url

	source, err := r.client.Do(r.req)

	if err != nil {
		return fmt.Errorf("Delete: error dispatching request %w", err)
	}

	res.WriteRequest(r.req)
	return HTTPRequestAdapterResponse(res, source, err)
}
