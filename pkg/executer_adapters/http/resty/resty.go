package restyadapter

import (
	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

// RestyHTTPClientAdapter implements executer's Client interface

type RestyHTTPClientAdapter struct {
	client *resty.Client
}

func NewHTTPClientAdapter() *RestyHTTPClientAdapter {
	return &RestyHTTPClientAdapter{
		client: resty.New(),
	}
}

func (r *RestyHTTPClientAdapter) Request() executorhttp.Requester {
	return RestyHTTPRequestAdapter{
		req: r.client.R(),
	}
}
