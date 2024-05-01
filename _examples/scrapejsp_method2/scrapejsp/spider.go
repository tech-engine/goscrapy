package scrapejsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tech-engine/goscrapy/cmd/gos"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

func NewSpider(ctx context.Context) (*Spider, <-chan error) {

	// use proxies
	// proxies := core.WithProxies("proxy_url1", "proxy_url2", ...)
	// core := gos.New[*Record]().WithClient(
	// 	gos.DefaultClient(proxies),
	// )

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

	req := s.NewRequest()
	// req.Meta("JOB", job)
	req.Url("https://jsonplaceholder.typicode.com/todos/1")

	s.Request(req, s.parse)
}

func (s *Spider) Close(ctx context.Context) {
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	fmt.Printf("status: %d", resp.StatusCode())

	var data Record
	err := json.Unmarshal(resp.Bytes(), &data)
	if err != nil {
		log.Fatalln(err)
	}

	// to push to pipelines
	s.Yield(&data)
}
