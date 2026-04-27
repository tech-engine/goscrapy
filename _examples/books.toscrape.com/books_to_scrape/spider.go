package books_to_scrape

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func (s *Spider) StartRequest(ctx context.Context, job *Job) {
	// for each request we must call NewRequest() and never reuse it
	req := s.Request(ctx).Url(s.baseUrl)

	// GET is the default method
	s.Parse(req, s.parse)
}

// can be called when spider is about to close
func (s *Spider) Close(ctx context.Context) {
	s.Logger().Info("closing")
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	s.Logger().Infof("GET: %d %s", resp.StatusCode(), resp.Request().URL.String())
	productUrls := resp.Css("article.product_pod h3 a").Attr("href")
	s.Logger().Debugf("Found %d product URLs", len(productUrls))
	for _, productUrl := range productUrls {
		if strings.HasPrefix(productUrl, "catalogue/") {
			productUrl = fmt.Sprintf("%s/%s", s.baseUrl, productUrl)
		} else {
			productUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, productUrl)
		}
		req := s.Request(ctx).Url(productUrl)
		s.Parse(req, s.parseProduct)
		s.Logger().Infof("GET: %s", productUrl)
	}

	// pagination
	nextPage := resp.Css("li.next a").Attr("href")
	if len(nextPage) > 0 {
		nextPageUrl := fmt.Sprintf("%s/%s", s.baseUrl, nextPage[0])
		if !strings.Contains(nextPage[0], "catalogue/") {
			nextPageUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, nextPage[0])
		}
		s.Parse(s.Request(ctx).Url(nextPageUrl), s.parse)
	}
}

func (s *Spider) parseProduct(ctx context.Context, resp core.IResponseReader) {
	product := resp.Css("article.product_page")

	var title string
	if titles := product.Css(".product_main h1").Text(); len(titles) > 0 {
		title = titles[0]
	}

	var price string
	if prices := product.Css(".price_color").Text(); len(prices) > 0 {
		price = prices[0]
	}

	var stock string
	if stocks := product.Css(".availability").Text(); len(stocks) > 0 {
		match := regexp.MustCompile(`\((\d+) available\)`).FindStringSubmatch(strings.TrimSpace(stocks[0]))

		if len(match) > 0 {
			stock = match[1]
		}
	}

	var rating string
	if ratingClassAttrs := product.Css(".star-rating").Attr("class"); len(ratingClassAttrs) > 0 {
		rating = strings.Split(ratingClassAttrs[0], " ")[1]

	}

	var productDescription string
	if productDescriptions := product.Css("#product_description + *").Text(); len(productDescriptions) > 0 {
		productDescription = productDescriptions[0]
	}

	var upc string
	if upcs := product.Css("table tr:nth-child(1) td").Text(); len(upcs) > 0 {
		upc = upcs[0]
	}

	var productType string
	if productTypes := product.Css("table tr:nth-child(2) td").Text(); len(productTypes) > 0 {
		productType = productTypes[0]
	}

	var reviewCount string
	if reviewCounts := product.Css("table tr:nth-child(7) td").Text(); len(reviewCounts) > 0 {
		reviewCount = reviewCounts[0]
	}

	s.Yield(&Record{
		Title:       title,
		Price:       price,
		Stock:       stock,
		Rating:      rating,
		Description: productDescription,
		Upc:         upc,
		ProductType: productType,
		Reviews:     reviewCount,
	})
}
