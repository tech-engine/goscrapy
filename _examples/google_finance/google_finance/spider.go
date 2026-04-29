// I obtained permission to re-create it on goscrapy.
// credits: https://scraper.run/blog/reverse-engineering-google-finance
package google_finance

import (
	"context"
	"fmt"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/gos"
)

const GOOGLE_RPC_URL = "https://www.google.com/finance/_/GoogleFinanceUi/data/batchexecute"

var RPC_HEADERS = map[string]string{
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Language": "en-US,en;q=0.9",
	"Accept-Encoding": "identity",
	"Cookie":          "CONSENT=YES+",
	"Content-Type":    "application/x-www-form-urlencoded;charset=UTF-8",
}

// StartRequest initiates the scraping process for a given ticker job.
func (s *Spider) StartRequest(ctx context.Context, job *Job) {
	ticker := job.Query()

	// build request body and retrieve required rpcids for the URL
	body, rpcids := buildBatchRequest(ticker)

	targetUrl := fmt.Sprintf("%s?rpcids=%s&source-path=/finance/quote/%s&hl=en&gl=us&rt=c",
		GOOGLE_RPC_URL, rpcids, ticker)

	req := s.Request(ctx).
		Url(targetUrl).
		Method("POST").
		Body(body).
		Meta("ticker", ticker)

	// apply required headers for Google Finance RPC
	for k, v := range RPC_HEADERS {
		req.SetHeader(k, v)
	}

	s.Parse(req, s.parseResponse)
}

// parseResponse processes the RPC response and maps the extracted data to a Record.
func (s *Spider) parseResponse(ctx context.Context, resp core.IResponseReader) {
	tickerRaw, _ := resp.Meta("ticker")
	ticker := tickerRaw.(string)

	// extract individual RPC results from the batch response
	results := parseBatchResults(resp.Bytes())

	record := &Record{J: NewJob(ticker)}

	// map extracted data parts into the final record structure
	mapQuote(ticker, results["xh8wxf"], record)
	mapCompany(results["HqGpWd"], record)
	mapChart(results["AiCwsd"], record)
	mapNews(results["nBEQBc"], record)

	// ensure we successfully captured the core quote data
	if record.Ticker == "" {
		s.Logger().Errorf("Failed to extract quote data for ticker: %s", ticker)
		return
	}

	s.Yield(record)
}
