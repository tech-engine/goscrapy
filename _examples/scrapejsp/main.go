package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"scrapejsp/scrapejsp"
	"sync"
	"syscall"

	// replace with your own project name

	"github.com/tech-engine/goscrapy/cmd/gos"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines"
)

// sample terminate function to demostrate spider termination.
func OnTerminate(fn func()) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	fn()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	// get core spider
	gos := gos.New[*scrapejsp.Record]()

	// use proxies
	// proxies := gos.WithProxies("proxy_url1", "proxy_url2", ...)

	// get core spider
	// gos := gos.New[*scrapejsp.Record]().WithClient(
	// 	gos.DefaultClient(proxies),
	// )

	// we can use middlewares like below
	// gos.MiddlewareManager.Add(
	// 	middlewares.DupeFilter,
	// 	middlewares.MultiCookieJar,
	// )

	export2Csv := pipelines.Export2CSV[*scrapejsp.Record]()
	export2Csv.WithFilename("itstimeitsnowornever.csv")

	// use export 2 json pipeline
	// export2Json := pipelines.Export2JSON[*scrapejsp.Record]()
	// export2Json.WithImmediate()
	// export2Json.WithFilename("itstimeitsnowornever.json")
	// we can use piplines
	gos.PipelineManager.Add(
		export2Csv,
		// export2Json,
	)

	go func() {
		defer wg.Done()

		err := gos.Start(ctx)

		if err != nil && errors.Is(err, context.Canceled) {
			return
		}

		fmt.Printf("failed: %q", err)
	}()

	spider := scrapejsp.NewSpider(gos)

	// start the scraper with a job, currently nil is passed but you can pass your job here
	spider.StartRequest(ctx, nil)

	OnTerminate(func() {
		fmt.Println("exit signal received: shutting down gracefully")
		cancel()
		wg.Wait()
	})

}
