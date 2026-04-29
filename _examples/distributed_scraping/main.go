package main

import (
	"context"
	"errors"
	"fmt"

	"distributed_scraping/distributed_scraping"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the distributed spider
	spider, err := distributed_scraping.New(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to create spider: %v\n", err)
		return
	}

	fmt.Println("🕷️  Distributed GoScrapy spider is running (Redis-backed).")

	// Start and wait for completion
	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("❌ Engine finished with error: %v\n", err)
	} else {
		fmt.Println("✨ Engine finished successfully.")
	}
}
