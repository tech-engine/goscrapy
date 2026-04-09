package fingerprint_spoofing

import (
	"context"

	"github.com/tech-engine/goscrapy/cmd/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

func New(ctx context.Context) (*Spider, <-chan error) {

	core := gos.New[*Record]().Setup(MIDDLEWARES, PIPELINES)

	errCh := make(chan error)

	spider := &Spider{
		core,
	}

	go func() {
		errCh <- core.Start(ctx)
		spider.Close(ctx)
	}()

	return spider, errCh
}
