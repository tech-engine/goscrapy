package main

import (
	"context"
	"errors"
	"fmt"

	// replace with your own project name
	"fingerprint_spoofing/fingerprint_spoofing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider, err := fingerprint_spoofing.New(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to create spider: %v\n", err)
		return
	}



	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion
	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("❌ Engine finished with error: %v\n", err)
	} else {
		fmt.Println("✨ Engine finished successfully.")
	}
}
