// Note: This scraper was created using goscrapy and for educational purposes only
// to showcase the capabilities of goscrapy and I am not liable for any misuse of this scraper.
package google_maps_scraper

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func (s *Spider) StartRequest(ctx context.Context, job *Job) {
	if job == nil {
		return
	}

	if job.loc == nil {
		s.Logger().Infof("Geocoding query: %s", job.query)
		req := prepareRequest(s.NewRequest(ctx), generateGeocodingUrl("https://www.google.com", *job), *job)
		s.Request(req, s.parseGeocoding)
		return
	}

	// call the next parse method
	req := prepareRequest(s.NewRequest(ctx), generateSearchUrl("https://www.google.com", *job), *job)
	s.Request(req, s.parseMapListing)
}

// can be called when spider is about to close
func (s *Spider) Close(ctx context.Context) {
	s.Logger().Info("Closing spider")
}

func (s *Spider) parseGeocoding(ctx context.Context, resp core.IResponseReader) {
	job, ok := getJob(resp)
	if !ok {
		return
	}

	lat, lng, _, ok := extractGeocoding(resp.Bytes())
	if !ok {
		return
	}

	job.setLocation(lat, lng)
	s.Logger().Infof("Location found for %s: %f, %f", job.query, lat, lng)

	req := prepareRequest(s.NewRequest(ctx), generateSearchUrl("https://www.google.com", job), job)
	s.Request(req, s.parseMapListing)
}

func (s *Spider) parseMapListing(ctx context.Context, resp core.IResponseReader) {
	job, ok := getJob(resp)
	if !ok {
		return
	}

	if resp.StatusCode() != 200 {
		return
	}

	records := extractMapResults(resp.Bytes())
	s.Logger().Infof("Found %d records for [%s] (At cursor %d)", len(records), job.query, job.cursor)
	if len(records) <= 0 {
		return
	}

	// update cursor
	job.SetCursor(job.cursor + 20)
	for _, record := range records {
		s.Yield(&record)
	}

	req := prepareRequest(s.NewRequest(ctx), generateSearchUrl("https://www.google.com", job), job)
	s.Request(req, s.parseMapListing)
}
