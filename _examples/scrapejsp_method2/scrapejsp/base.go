package scrapejsp

import (
	"context"

	"github.com/tech-engine/goscrapy/cmd/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

func New(ctx context.Context) (*Spider, <-chan error) {
	app := gos.NewApp[*Record]().
		Setup(MIDDLEWARES, PIPELINES)

	spider := &Spider{
		ICoreSpider: app,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Start(ctx)
		spider.Close(ctx)
	}()

	return spider, errCh
}
