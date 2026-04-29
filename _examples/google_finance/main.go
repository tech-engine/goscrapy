package main

import (
	"context"
	"errors"
	"fmt"

	"google_finance/google_finance"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spider, err := google_finance.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create spider: %v\n", err)
		return
	}

	// our query
	ticker := "GOOGL:NASDAQ"

	// create a job with the ticker
	job := google_finance.NewJob(ticker)

	// start the request with the job
	spider.StartRequest(ctx, job)

	if err := spider.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("Engine finished with error: %v\n", err)
	} else {
		fmt.Println("Engine finished successfully.")
	}
}
