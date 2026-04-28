package use_signals

import (
	"context"
	"log"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

// New initializes spider
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

	// explicit signal subscription
	app.OnEngineStarted(func(ctx context.Context) {
		log.Println("[signal] engine started")
	}).
		OnRequestScheduled(func(ctx context.Context, req *core.Request) {
			log.Printf("[signal] request scheduled: %s", req.URL)
		}).
		OnResponseReceived(func(ctx context.Context, resp core.IResponseReader) {
			log.Printf("[signal] response received: %s (status: %d)", resp.Request().URL, resp.StatusCode())
		}).
		OnItemScraped(spider.onItemScraped).
		OnRequestDropped(spider.onRequestDropped).
		OnItemDropped(spider.onItemDropped).
		OnEngineStopped(func(ctx context.Context) {
			log.Println("[signal] engine stopped")
		})

	// RegisterSpider triggers Auto-Discovery of spider hooks (Open/Close/Idle/Error)
	app.RegisterSpider(spider)

	go func() {
		_ = app.Start(ctx)
	}()

	return spider, nil
}
