package main

import (
	"context"
	"errors"
	"fmt"

	// replace with your own project name
	"with_tui_stats/with_tui_stats"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider, err := with_tui_stats.New(ctx, true)
	if err != nil {
		fmt.Printf("❌ Failed to create spider: %v\n", err)
		return
	}

	// start the scraper with a job, currently nil is passed but you can pass your job here
	spider.StartRequest(ctx, nil)

	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion
	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("❌ Engine finished with error: %v\n", err)
	} else {
		fmt.Println("✨ Engine finished successfully.")
	}
}
