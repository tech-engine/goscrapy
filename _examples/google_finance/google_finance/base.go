package google_finance

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

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

	app.RegisterSpider(spider)

	go func() {
		_ = app.Start(ctx)
	}()

	return spider, nil
}
