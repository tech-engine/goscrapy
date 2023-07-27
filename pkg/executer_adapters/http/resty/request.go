package restyadapter

import (
	"context"

	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// RestyHTTPRequestAdapter implements executer's Requester interface

type RestyHTTPRequestAdapter struct {
	req *resty.Request
}

func (r RestyHTTPRequestAdapter) SetContext(ctx context.Context) executorhttp.Requester {
	r.req.SetContext(ctx)
	return r
}

func (r RestyHTTPRequestAdapter) SetHeaders(headers map[string]string) executorhttp.Requester {
	r.req.SetHeaders(headers)
	return r
}

func (r RestyHTTPRequestAdapter) SetBody(body any) executorhttp.Requester {
	r.req.SetBody(body)
	return r
}

func (r RestyHTTPRequestAdapter) Get(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Get(url)
	return RestyHTTPRequestAdapterResponse(target, source, err)
}

func (r RestyHTTPRequestAdapter) Post(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Post(url)
	return RestyHTTPRequestAdapterResponse(target, source, err)
}

func (r RestyHTTPRequestAdapter) Put(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Put(url)
	return RestyHTTPRequestAdapterResponse(target, source, err)
}

func (r RestyHTTPRequestAdapter) Patch(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Patch(url)
	return RestyHTTPRequestAdapterResponse(target, source, err)
}

func (r RestyHTTPRequestAdapter) Delete(target executorhttp.ResponseWriter, url string) error {
	source, err := r.req.Delete(url)
	return RestyHTTPRequestAdapterResponse(target, source, err)
}
