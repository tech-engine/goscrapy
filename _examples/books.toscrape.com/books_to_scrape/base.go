package books_to_scrape

import (
	"context"

	"github.com/tech-engine/goscrapy/cmd/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

// New initializes the spider with minimal setup (no TUI, no stats collection).
func New(ctx context.Context) (*Spider, <-chan error) {
	app := gos.NewApp[*Record]().
		Setup(MIDDLEWARES, PIPELINES)

	spider := &Spider{
		ICoreSpider: app,
		baseUrl:     "https://books.toscrape.com",
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Start(ctx)
		spider.Close(ctx)
	}()

	return spider, errCh
}
