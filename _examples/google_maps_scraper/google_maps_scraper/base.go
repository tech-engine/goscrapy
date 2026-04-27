package google_maps_scraper

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

// New initializes the spider with the new v0.20.1 pattern.
func New(ctx context.Context) (*Spider, error) {
	app, err := gos.New[*Record]()
	if err != nil {
		return nil, err
	}

	app.WithMiddlewares(MIDDLEWARES...).
		WithPipelines(PIPELINES...)

	spider := &Spider{
		ICoreSpider: app,
	}

	go func() {
		_ = app.Start(ctx)
	}()

	return spider
}
