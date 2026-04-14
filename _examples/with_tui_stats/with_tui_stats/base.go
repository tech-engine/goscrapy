package with_tui_stats

import (
	"context"

	"github.com/tech-engine/goscrapy/cmd/gos"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
	"github.com/tech-engine/goscrapy/pkg/tui"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

// New initializes the spider with optional TUI and stats collection enabled.
func New(ctx context.Context, tuiEnabled bool) (*Spider, <-chan error) {
	app := gos.NewApp[*Record]().
		Setup(MIDDLEWARES, PIPELINES)

	spider := &Spider{
		ICoreSpider: app,
		baseUrl:     "https://books.toscrape.com",
	}

	errCh := make(chan error, 1)

	if tuiEnabled {
		// Explicitly wire the user stats collector to framework engines
		app.Scheduler.WithStatsRecorderFactory(HttpStats)

		// Configure Telemetry
		hub := ts.NewTelemetryHub()
		hub.AddCollector(HttpStats)

		app.Engine.WithOnShutdown(func() {
			// Output stats cleanly at the end
			// We can have any other cleanup code here
			HttpStats.Print()
		})

		dashboard := tui.New(app.Logger())
		hub.AddObserver(dashboard)
		app.WithTelemetry(hub)

		go func() {
			errCh <- gos.StartWithTUI(ctx, app, dashboard)
			spider.Close(ctx) // Optional finalizer for spider
		}()
	} else {
		go func() {
			errCh <- app.Start(ctx)
			spider.Close(ctx) // Optional finalizer for spider
		}()
	}

	return spider, errCh
}
