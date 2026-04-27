package with_tui_stats

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
	"github.com/tech-engine/goscrapy/pkg/signal"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
	"github.com/tech-engine/goscrapy/pkg/tui"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

// New initializes the spider with optional TUI and stats collection enabled.
func New(ctx context.Context, tuiEnabled bool) (*Spider, error) {
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

	if tuiEnabled {
		// Explicitly wire the user stats collector to framework engines
		app.WithStatsRecorderFactory(HttpStats)

		// Configure Telemetry
		hub := ts.NewTelemetryHub()
		hub.AddCollector(HttpStats)

		app.AddSignal(signal.EngineStopped, func(ctx context.Context) {
			// Output stats cleanly at the end
			// We can have any other cleanup code here
			HttpStats.Print()
		})

		dashboard := tui.New(app.Logger())
		hub.AddObserver(dashboard)
		app.WithTelemetry(hub)

		go func() {
			_ = gos.StartWithTUI(ctx, app, dashboard)
		}()
	} else {
		go func() {
			_ = app.Start(ctx)
		}()
	}

	return spider
}
