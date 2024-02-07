package scrapejsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tech-engine/goscrapy/cmd/corespider"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type Spider struct {
	corespider.ICoreSpider[[]Record]
}

func NewSpider(core corespider.ICoreSpider[[]Record]) *Spider {
	return &Spider{
		core,
	}
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

func (s *Spider) parse(ctx context.Context, response core.IResponseReader) {
	fmt.Printf("status: %d", res.StatusCode())

	var data Record
	err := json.Unmarshal(res.Bytes(), &data)
	if err != nil {
		log.Fatalln(err)
	}

	// job, _ := res.Meta("JOB")
	output := &Output{
		records: []Record{data},
	}

	// to push to pipelines
	s.Yield(output)
}
