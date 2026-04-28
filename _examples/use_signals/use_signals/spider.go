package use_signals

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func (s *Spider) Open(ctx context.Context) {
	log.Println("[spider hook] open called")
	url := "https://jsonplaceholder.typicode.com/todos/1"

	log.Println("Scheduling first request...")
	req1 := s.Request(ctx).Url(url)
	s.Parse(req1, s.parse)

	log.Println("scheduling duplicate request to trigger dropped...")
	req2 := s.Request(ctx).Url(url)
	s.Parse(req2, s.parse)
}

func (s *Spider) Close(ctx context.Context) {
	log.Println("[spider hook] close called")
}

func (s *Spider) Idle(ctx context.Context) {
	log.Println("[spider hook] idle called")
}

func (s *Spider) Error(ctx context.Context, err error) {
	log.Printf("[spider hook] error called: %v", err)
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	log.Printf("[spider] parsing response from %s", resp.Request().URL)

	var data Record
	err := json.Unmarshal(resp.Bytes(), &data)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		return
	}

	log.Printf("[spider] scraped title: %s", data.Title)
	s.Yield(&data)
}

func (s *Spider) onItemScraped(ctx context.Context, item *Record) {
	log.Printf("[spider signal] item scraped: %s", item.Title)
}

func (s *Spider) onRequestDropped(ctx context.Context, req *core.Request, err error) {
	log.Printf("[spider signal] request dropped: %s (reason: %v)", req.URL, err)
}

func (s *Spider) onItemDropped(ctx context.Context, item *Record, err error) {
	log.Printf("[spider signal] item dropped: %s (reason: %v)", item.Title, err)
}
