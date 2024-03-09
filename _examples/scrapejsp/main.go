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

	// use middlewares
	gos.MiddlewareManager.Add(scrapejsp.MIDDLEWARES...)

	// use pipelines
	gos.PipelineManager.Add(scrapejsp.PIPELINES...)

	// export2Csv := pipelines.Export2CSV[*scrapejsp.Record](pipelines.Export2CSVOpts{
	// 	Filename: "itstimeitsnowornever.csv",
	// })

	// // use export 2 json pipeline
	// export2Json := pipelines.Export2JSON[*scrapejsp.Record](pipelines.Export2JSONOpts{
	// 	Filename:  "itstimeitsnowornever.json",
	// 	Immediate: true,
	// })

	// add pipeline to group
	// pipelineGroup := pm.NewGroup[*scrapejsp.Record]()
	// pipelineGroup.Add(export2Csv)
	// pipelineGroup.Add(export2Json)
	// gos.PipelineManager.Add(
	// 	pipelineGroup,
	// )

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
