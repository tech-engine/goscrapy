package google_maps_scraper

import (
	"context"

	"github.com/tech-engine/goscrapy/cmd/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

// New initializes the spider with the new v0.20.1 pattern.
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
