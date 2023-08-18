package restyadapter

import (
	"context"
	"io"

	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// RestyHTTPRequestAdapter implements executer's Requester interface

type HTTPRequestAdapter struct {
	req *resty.Request
}

func (r HTTPRequestAdapter) SetContext(ctx context.Context) executorhttp.Requester {
	r.req.SetContext(ctx)
	return r
}

func (r HTTPRequestAdapter) SetHeaders(headers map[string]string) executorhttp.Requester {
	r.req.SetHeaders(headers)
	return r
}

func (r HTTPRequestAdapter) SetBody(body io.ReadCloser) executorhttp.Requester {
	r.req.SetBody(body)
	return r
}

func (r HTTPRequestAdapter) Get(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Get(url)
	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Post(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Post(url)
	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Put(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Put(url)
	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Patch(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Patch(url)
	return HTTPRequestAdapterResponse(target, source, err)
}

func (r HTTPRequestAdapter) Delete(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Delete(url)
	return HTTPRequestAdapterResponse(target, source, err)
}
