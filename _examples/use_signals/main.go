package main

import (
	"context"
	"errors"
	"fmt"

	"os"

	// replace with your own project name
	"use_signals/use_signals"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider, err := use_signals.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create spider: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion
	// By default autoExit is set to false, which means the engine will shut down
	// automatically when all work is finished.
	// If you want the engine to continue running indefinitely (e.g., if it's
	// accepting jobs from an external source), don't pass true to Wait().
	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "❌ Engine finished with error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✨ Engine finished successfully.")
}
