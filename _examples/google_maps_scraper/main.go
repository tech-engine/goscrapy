// Note: This scraper was created using goscrapy and for educational purposes only
// to showcase the capabilities of goscrapy and I am not liable for any misuse of this scraper.
package main

import (
	"context"
	"errors"
	"fmt"

	// replace with your own project name
	"google_maps_scraper/google_maps_scraper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start spider
	spider, errCh := google_maps_scraper.New(ctx)

	// create a job
	job := google_maps_scraper.NewJob("googlemaps_carwash")
	job.SetQuery("car wash in CO, USA").SetMaxRecords(60)

	// start the scraper with a job
	spider.StartRequest(ctx, job)

	fmt.Println("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")

	// wait for completion

	if err := spider.Wait(cancel, errCh); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("❌ Engine finished with error: %v\n", err)
	} else {
		fmt.Println("✨ Engine finished successfully.")
	}
}
