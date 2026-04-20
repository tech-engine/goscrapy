package with_tui_stats

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
	"github.com/tech-engine/goscrapy/pkg/tui"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

// New initializes the spider with optional TUI and stats collection enabled.
func New(ctx context.Context, tuiEnabled bool) *Spider {
	app := gos.NewApp[*Record]().
		WithMiddlewares(MIDDLEWARES...).
		WithPipelines(PIPELINES...)

	spider := &Spider{
		ICoreSpider: app,
		baseUrl:     "https://books.toscrape.com",
	}

	if tuiEnabled {
		// Explicitly wire the user stats collector to framework engines
		app.WithStatsRecorderFactory(HttpStats)

		// Configure Telemetry
		hub := ts.NewTelemetryHub()
		hub.AddCollector(HttpStats)

		app.WithOnEngineShutdown(func() {
			// Output stats cleanly at the end
			// We can have any other cleanup code here
			HttpStats.Print()
		})

		dashboard := tui.New(app.Logger())
		hub.AddObserver(dashboard)
		app.WithTelemetry(hub)

		go func() {
			_ = gos.StartWithTUI(ctx, app, dashboard)
			spider.Close(ctx) // Optional finalizer for spider
		}()
	} else {
		go func() {
			_ = app.Start(ctx)
			spider.Close(ctx) // Optional finalizer for spider
		}()
	}

	return spider
}
