package scrapeThisSite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func NewSpider() (*Spider, error) {
	// create pool
	var (
		err  error
		_url *url.URL
		s    *Spider
	)

	_url, err = url.Parse("https://www.scrapethissite.com")

	if err != nil {
		return nil, err
	}

	s = &Spider{
		baseUrl: _url,
	}

	return s, nil
}

// This is the entrypoint to the spider
func (s *Spider) StartRequest(ctx context.Context, job *Job) {

	var headers map[string]string
	years := []string{"2010", "2011", "2012", "2013", "2014", "2015"}

	// GET is the default request method
	for _, year := range years {
		// for each request we must call NewRequest() and never reuse it
		req := s.NewRequest()

		req.SetUrl(s.baseUrl.String()+job.query+year).
			SetMetaData("JOB", job).
			SetHeaders(headers)

		// call the next parse method
		s.Request(ctx, req, s.parse)
	}
}

func (s *Spider) parse(ctx context.Context, response core.ResponseReader) {

	var (
		err error
		job *Job
	)

	output := &Output{}

	defer func() {
		if err != nil {
			output.err = err
		}
		s.yield(output)
	}()

	job, err = s.jobFromContext(ctx)

	if err != nil {
		output.err = err
		return
	}

	var data []Record
	err = json.Unmarshal(response.Body(), &data)

	if err != nil {
		fmt.Print("Error", err)
		return
	}

	output.records = data
	output.job = job

}
