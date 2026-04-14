package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	// replace with your own project name
	"books_to_scrape/books_to_scrape"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider := books_to_scrape.New(ctx)

	// start the scraper with a job, currently nil is passed but you can pass your job here
	spider.StartRequest(ctx, nil)

	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion
	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "❌ Engine finished with error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✨ Engine finished successfully.")
}
