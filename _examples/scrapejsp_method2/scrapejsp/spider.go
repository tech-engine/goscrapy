package scrapejsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// open is auto-called by goscrapy during engine startup
func (s *Spider) Open(ctx context.Context) {
	req := s.Request(ctx).
		Url("https://jsonplaceholder.typicode.com/todos/1")
	s.Parse(req, s.parse)
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
