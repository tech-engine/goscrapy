package fingerprint_spoofing

import (
	"context"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/core"
)

// open is auto-called by goscrapy during engine startup
func (s *Spider) Open(ctx context.Context) {
	ctx = middlewares.WithAzureTLSOptions(ctx, &middlewares.AzureTLSOptions{
		Browser: middlewares.BrowserFirefox,
		Proxy:   "http://user:pass@myproxy.com:8080",
	})
	req := s.Request(ctx).Url("https://tls.peet.ws/api/all")
	s.Parse(req, s.parse)
}

func (s *Spider) Close(ctx context.Context) {
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	fmt.Printf("status: %d\n", resp.StatusCode())

	fmt.Println(string(resp.Bytes()))
}
