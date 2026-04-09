package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/tech-engine/goscrapy/cmd/gos"
	// replace with your own project name
	"scrapejsp/scrapejsp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider, errCh := scrapejsp.New(ctx)

	// start the scraper with a job, currently nil is passed but you can pass your job here
	spider.StartRequest(ctx, nil)

	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion
	if err := gos.Wait(cancel, errCh); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("❌ Engine finished with error: %v\n", err)
	} else {
		fmt.Println("✨ Engine finished successfully.")
	}
}
