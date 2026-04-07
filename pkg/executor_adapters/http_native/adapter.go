package httpnative

import (
	"fmt"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

// HTTPAdapter implements Executor's ExecAdapter interface
type HTTPAdapter struct {
	client *http.Client
}

func NewHTTPClientAdapter(client *http.Client, poolSize uint64) *HTTPAdapter {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPAdapter{
		client: client,
	}
}



func (r *HTTPAdapter) WithClient(client *http.Client) {
	r.client = client
}

func (r *HTTPAdapter) Do(res engine.IResponseWriter, req *http.Request) error {
	source, err := r.client.Do(req)

	if err != nil {
		return fmt.Errorf("Do: error dispatching request %w", err)
	}

	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
