package with_tui_stats

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func (s *Spider) StartRequest(ctx context.Context, job *Job) {

	// for each request we must call NewRequest() and never reuse it
	req := s.NewRequest(ctx)

	// GET is the default method
	req.Url(s.baseUrl)

	s.Request(req, s.parse)
}

// can be called when spider is about to close
func (s *Spider) Close(ctx context.Context) {
	s.Logger().Info("closing")
}

func (s *Spider) parse(ctx context.Context, resp core.IResponseReader) {
	s.Logger().Infof("GET: %d %s", resp.StatusCode(), resp.Request().URL.String())
	for _, productUrl := range resp.Css("article.product_pod h3 a").Attr("href") {
		req := s.NewRequest(ctx)

		if strings.HasPrefix(productUrl, "catalogue/") {
			productUrl = fmt.Sprintf("%s/%s", s.baseUrl, productUrl)
		} else {
			productUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, productUrl)
		}

		req.Url(productUrl)
		s.Request(req, s.parseProduct)
		s.Logger().Infof("GET: %s", productUrl)
	}

	// pagination
	nextUrls := resp.Css("li.next a").Attr("href")

	if len(nextUrls) <= 0 {
		return
	}

	nextUrl := fmt.Sprintf("%s/%s", s.baseUrl, nextUrls[0])

	if !strings.HasPrefix(nextUrls[0], "catalogue/") {
		nextUrl = fmt.Sprintf("%s/catalogue/%s", s.baseUrl, nextUrls[0])
	}

	req := s.NewRequest(ctx)
	req.Url(nextUrl)
	s.Request(req, s.parse)
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
