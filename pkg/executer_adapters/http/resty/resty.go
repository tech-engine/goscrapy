package restyadapter

import (
	"net/http"

	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// RestyHTTPClientAdapter implements executer's Client interface

type HTTPClientAdapter struct {
	client *resty.Client
}

func NewHTTPClientAdapter(client *http.Client) *HTTPClientAdapter {
	if client == nil {
		client = &http.Client{}
	}
	return &HTTPClientAdapter{
		client: resty.NewWithClient(client),
	}
}

func (r *HTTPClientAdapter) Request() executorhttp.Requester {
	return HTTPRequestAdapter{
		req: r.client.R(),
	}
}
