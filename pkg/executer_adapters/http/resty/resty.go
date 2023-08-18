package restyadapter

import (
	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// RestyHTTPClientAdapter implements executer's Client interface

type HTTPClientAdapter struct {
	client *resty.Client
}

func NewHTTPClientAdapter() *HTTPClientAdapter {
	return &HTTPClientAdapter{
		client: resty.New(),
	}
}

func (r *HTTPClientAdapter) Request() executorhttp.Requester {
	return HTTPRequestAdapter{
		req: r.client.R(),
	}
}
