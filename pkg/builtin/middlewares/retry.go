package middlewares

import (
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

const MIDDLEWARE_HTTP_RETRY_MAX_RETRIES = 3

var MIDDLEWARE_HTTP_RETRY_CODES = []int{500, 502, 503, 504, 522, 524, 408, 429}

type RetryCb func(*http.Request, uint8) bool

type RetryOpts struct {
	MaxRetries uint8
	Codes      []int
	BaseDelay  time.Duration
	Cb         RetryCb
}

func defaultOpts() *RetryOpts {
	opts := &RetryOpts{
		MaxRetries: MIDDLEWARE_HTTP_RETRY_MAX_RETRIES,
		Codes:      MIDDLEWARE_HTTP_RETRY_CODES,
		BaseDelay:  1 * time.Second,
	}

	value, ok := os.LookupEnv("MIDDLEWARE_HTTP_RETRY_MAX_RETRIES")

	if ok {
		maxRetries, err := strconv.Atoi(value)
		if err == nil {
			opts.MaxRetries = uint8(maxRetries)
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_RETRY_CODES")

	if ok {
		codesStr := strings.Split(value, ",")
		codes := make([]int, 0, len(codesStr))

		for _, codeStr := range codesStr {
			if code, err := strconv.Atoi(strings.TrimSpace(codeStr)); err == nil {
				codes = append(codes, code)
			}
		}

		if len(codes) > 0 {
			opts.Codes = codes[:]
		}

	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_RETRY_BASE_DELAY")

	if ok {
		baseDelay, err := time.ParseDuration(value)
		if err == nil {
			opts.BaseDelay = baseDelay
		}
	}

	return opts
}

func Retry(opts RetryOpts) func(http.RoundTripper) http.RoundTripper {

	retryOpts := defaultOpts()

	if opts.MaxRetries > 0 {
		retryOpts.MaxRetries = opts.MaxRetries
	}

	if opts.Codes != nil {
		retryOpts.Codes = opts.Codes[:]
	}

	if opts.BaseDelay > 0 {
		retryOpts.BaseDelay = opts.BaseDelay
	}

	if opts.Cb != nil {
		retryOpts.Cb = opts.Cb
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {

			var (
				resp    *http.Response
				err     error
				retries uint8 = retryOpts.MaxRetries
				i       uint8
			)

			retryHeader := req.Header.Get("X-Goscrapy-Middleware-Max-Retry")

			if retryHeader != "" {
				r, _ := strconv.Atoi(retryHeader)
				retries = uint8(r)
				req.Header.Del("X-Goscrapy-Middleware-Max-Retry")
			}

			retries += 1

			timer := time.NewTimer(retryOpts.BaseDelay)

			for i = 0; i < retries; i++ {
				resp, err = next.RoundTrip(req)

				// call retry callback, if present
				if i > 0 && retryOpts.Cb != nil && !retryOpts.Cb(req, i) {
					break
				}

				if err != nil {
					select {
					case <-timer.C:
						// calculate next delay
						timer.Reset(time.Duration(math.Pow(2, float64(i))) * retryOpts.BaseDelay)
						continue
					}
				}

				if !slices.Contains(retryOpts.Codes, resp.StatusCode) {
					break
				}

				select {
				case <-timer.C:
					// calculate next delay
					timer.Reset(time.Duration(math.Pow(2, float64(i))) * retryOpts.BaseDelay)
				}
			}

			if !timer.Stop() {
				<-timer.C
			}

			return resp, err
		})
	}
}
