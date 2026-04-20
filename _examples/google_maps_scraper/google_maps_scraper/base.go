package google_maps_scraper

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

// New initializes the spider with the new v0.20.1 pattern.
func New(ctx context.Context) *Spider {
	app := gos.NewApp[*Record]().
		WithMiddlewares(MIDDLEWARES...).
		WithPipelines(PIPELINES...)

	spider := &Spider{
		ICoreSpider: app,
	}

	go func() {
		_ = app.Start(ctx)
		spider.Close(ctx)
	}()

	return spider
}
