package books_to_scrape

import (
	"context"
	"fmt"
	"strings"

	"github.com/tech-engine/goscrapy/pkg/builtin/gosm"
	"github.com/tech-engine/goscrapy/pkg/core"
)

// can be called when spider is about to close
func (s *Spider) Close(ctx context.Context) {
	s.Logger().Info("closing")
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	s.Logger().Infof("GET: %d %s", resp.StatusCode(), resp.Request().URL.String())

	var list Listing
	_ = gosm.Map(resp, &list)

	s.Logger().Debugf("Found %d product URLs", len(list.ProductLinks))
	for _, productUrl := range list.ProductLinks {
		if strings.HasPrefix(productUrl, "catalogue/") {
			productUrl = fmt.Sprintf("%s/%s", s.baseUrl, productUrl)
		} else {
			productUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, productUrl)
		}
		s.Parse(s.Request(ctx).Url(productUrl), s.parseProduct)
		s.Logger().Infof("GET: %s", productUrl)
	}

	// pagination
	if list.NextPage != "" {
		nextPageUrl := fmt.Sprintf("%s/%s", s.baseUrl, list.NextPage)
		if !strings.Contains(list.NextPage, "catalogue/") {
			nextPageUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, list.NextPage)
		}
		s.Parse(s.Request(ctx).Url(nextPageUrl), s.parse)
	}
}

func (s *Spider) parseProduct(ctx context.Context, resp core.IResponseReader) {
	record := &Record{}
	_ = gosm.Map(resp, record)

	// Rating: gosm extracts raw class ("star-rating Three"), extract word
	if parts := strings.Split(record.Rating, " "); len(parts) > 1 {
		record.Rating = parts[1]
	}

	s.Yield(record)
}

func (s *Spider) StartRequests(ctx context.Context) {
	s.Parse(s.Request(ctx).Url(s.baseUrl), s.parse)
}
