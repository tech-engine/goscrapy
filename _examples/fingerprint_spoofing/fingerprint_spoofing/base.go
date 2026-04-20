package fingerprint_spoofing

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
}

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
