package httpnative

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

const EX_ADAPTER_DEFAULT_REQ_RES_POOL_SIZE = 1e6

// HTTPAdapter implements Executor's ExecAdapter interface
type HTTPAdapter struct {
	client  *http.Client
	reqpool *rp.Pooler[http.Request]
}

func NewHTTPClientAdapter(client *http.Client, poolSize uint64) *HTTPAdapter {
	if client == nil {
		client = http.DefaultClient
	}

	if poolSize == 0 {
		poolSize = EX_ADAPTER_DEFAULT_REQ_RES_POOL_SIZE
		value, ok := os.LookupEnv("SCHEDULER_DEFAULT_REQ_RES_POOL_SIZE")

		if ok {
			parsedPoolSize, err := strconv.ParseUint(value, 10, 64)
			if err == nil {
				poolSize = parsedPoolSize
			}
		}
	}

	return &HTTPAdapter{
		client:  client,
		reqpool: rp.NewPooler[http.Request](rp.WithSize[http.Request](poolSize)),
	}
}

func (r *HTTPAdapter) Acquire() *http.Request {
	req := r.reqpool.Acquire()
	if req == nil {
		req = &http.Request{}
	}
	return req
}

func (r *HTTPAdapter) WithClient(client *http.Client) {
	r.client = client
}

func (r *HTTPAdapter) Do(res engine.IResponseWriter, req *http.Request) error {
	defer r.reqpool.Release(req)

	source, err := r.client.Do(req)

	if err != nil {
		return fmt.Errorf("Do: error dispatching request %w", err)
	}

	res.WriteRequest(req)
	HTTPRequestAdapterResponse(res, source)
	return nil
}
