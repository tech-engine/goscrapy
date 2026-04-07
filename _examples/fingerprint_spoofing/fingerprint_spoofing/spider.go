package fingerprint_spoofing

import (
	"context"
	"fmt"

	"github.com/tech-engine/goscrapy/cmd/gos"
	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

func New(ctx context.Context) (*Spider, <-chan error) {

	core := gos.New[*Record]()

	// Add middlewares
	core.MiddlewareManager.Add(MIDDLEWARES...)
	// Add pipelines
	core.PipelineManager.Add(PIPELINES...)

	errCh := make(chan error)

	go func() {
		errCh <- core.Start(ctx)
	}()

	return &Spider{
		core,
	}, errCh
}

// This is the entrypoint to the spider
func (s *Spider) StartRequest(ctx context.Context, job *Job) {
	ctx = middlewares.WithAzureTLSOptions(ctx, &middlewares.AzureTLSOptions{
		Browser: middlewares.BrowserFirefox,
		Proxy:   "http://user:pass@myproxy.com:8080",
	})
	req := s.NewRequest(ctx)
	req.Url("https://tls.peet.ws/api/all")
	s.Request(req, s.parse)
}

func (s *Spider) Close(ctx context.Context) {
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	fmt.Printf("status: %d\n", resp.StatusCode())

	fmt.Println(string(resp.Bytes()))
}
