package books_to_scrape

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

// New initializes the spider with minimal setup (no TUI, no stats collection).
func New(ctx context.Context) (*Spider, error) {
	app, err := gos.New[*Record]()
	if err != nil {
		return nil, err
	}

	app.WithMiddlewares(MIDDLEWARES...).
		WithPipelines(PIPELINES...)

	spider := &Spider{
		ICoreSpider: app,
		baseUrl:     "https://books.toscrape.com",
	}

	go func() {
		_ = app.Start(ctx)
	}()

	return spider, nil
}
