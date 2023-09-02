package main

import (
	"context"
	"log"
	scrapeThisSite "scrapeThisSiteExample/scrapethissite"
	"scrapeThisSiteExample/utils"

	"github.com/tech-engine/goscrapy/pkg/pipelines"
)

func main() {
	/** Add comment according to you*/
	ctx, cancel := context.WithCancel(context.Background())

	goScrapy, err := scrapeThisSite.Setup(ctx)

	if err != nil {
		log.Fatalln(err)
	}

	// Using JSON pipeline to export data as JSON.
	goScrapy.Pipelines().Add(pipelines.Export2JSON[*scrapeThisSite.Job, []scrapeThisSite.Record]())

	if err := goScrapy.Start(ctx); err != nil {
		log.Fatalln(err)
	}

	// Create a new Spider Job
	spiderJob := goScrapy.NewJob("scrapeThisSite")

	// Set Job parameters
	spiderJob.SetQuery("/pages/ajax-javascript/?ajax=true&year=")

	// Run the spider
	goScrapy.Run(spiderJob)

	utils.OnTerminate(func() {
		cancel()
		// We'll wait until all go routines exit.
		goScrapy.Wait()
	})
}
