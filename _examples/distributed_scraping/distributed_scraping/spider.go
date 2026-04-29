package distributed_scraping

import (
	"context"
	"fmt"
	"strings"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/gos"
)

type Spider struct {
	gos.ICoreSpider[*Record]
	baseUrl string
}

func (s *Spider) Open(ctx context.Context) {
	s.Logger().Info("Spider opened, seeding initial request...")
	req := s.Request(ctx).Url(s.baseUrl)
	s.Parse(req, s.parse)
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	s.Logger().Infof("Parsing page: %s", resp.Request().URL.String())

	titles := resp.Css("article.product_pod h3 a").Attr("title")
	prices := resp.Css("article.product_pod .price_color").Text()

	for i := 0; i < len(titles) && i < len(prices); i++ {
		s.Yield(&Record{
			Title: titles[i],
			Price: prices[i],
		})
	}

	// Pagination
	for _, nextUrl := range resp.Css("li.next a").Attr("href") {
		if !strings.Contains(nextUrl, "catalogue/") {
			nextUrl = "catalogue/" + nextUrl
		}
		fullUrl := fmt.Sprintf("https://books.toscrape.com/%s", nextUrl)
		req := s.Request(ctx).Url(fullUrl)
		s.Parse(req, s.parse)
	}
}
