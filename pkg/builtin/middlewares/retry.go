package middlewares

import (
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

const MIDDLEWARE_HTTP_RETRY_MAX_RETRIES = 3

var MIDDLEWARE_HTTP_RETRY_CODES = []int{500, 502, 503, 504, 522, 524, 408, 429}

type RetryCb func(*http.Request, uint8) bool

type retryOpts struct {
	maxRetries uint8
	codes      []int
	baseDelay  time.Duration
	cb         RetryCb
}

func WithMaxRetries(max uint8) types.OptFunc[retryOpts] {
	return func(opts *retryOpts) {
		opts.maxRetries = max
	}
}

func WithHttpCodes(codes []int) types.OptFunc[retryOpts] {
	return func(opts *retryOpts) {
		opts.codes = codes
	}
}

func WithBaseDelay(dur time.Duration) types.OptFunc[retryOpts] {
	return func(opts *retryOpts) {
		opts.baseDelay = dur
	}
}

func WithCb(cb RetryCb) types.OptFunc[retryOpts] {
	return func(opts *retryOpts) {
		opts.cb = cb
	}
}

func DefaultRetryOpts() *retryOpts {
	opts := &retryOpts{
		maxRetries: MIDDLEWARE_HTTP_RETRY_MAX_RETRIES,
		codes:      MIDDLEWARE_HTTP_RETRY_CODES,
		baseDelay:  1 * time.Second,
	}

	value, ok := os.LookupEnv("MIDDLEWARE_HTTP_RETRY_MAX_RETRIES")

	if ok {
		maxRetries, err := strconv.Atoi(value)
		if err == nil {
			opts.maxRetries = uint8(maxRetries)
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
			opts.codes = codes[:]
		}

	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_RETRY_BASE_DELAY")

	if ok {
		baseDelay, err := time.ParseDuration(value)
		if err == nil {
			opts.baseDelay = baseDelay
		}
	}

	return opts
}

func Retry(optFns ...types.OptFunc[retryOpts]) func(http.RoundTripper) http.RoundTripper {

	retryOpts := DefaultRetryOpts()

	for _, optFn := range optFns {
		optFn(retryOpts)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {

			var (
				resp    *http.Response
				err     error
				retries uint8 = retryOpts.maxRetries
				i       uint8
			)

			retryHeader := req.Header.Get("X-Goscrapy-Middleware-Max-Retry")

			if retryHeader != "" {
				r, _ := strconv.Atoi(retryHeader)
				retries = uint8(r)
				req.Header.Del("X-Goscrapy-Middleware-Max-Retry")
			}

			retries += 1

			timer := time.NewTimer(retryOpts.baseDelay)

			for i = 0; i < retries; i++ {
				resp, err = next.RoundTrip(req)

				// call retry callback, if present
				if i > 0 && retryOpts.cb != nil && !retryOpts.cb(req, i) {
					break
				}

				if err != nil {
					select {
					case <-timer.C:
						// calculate next delay
						timer.Reset(time.Duration(math.Pow(2, float64(i))) * retryOpts.baseDelay)
						continue
					}
				}

				if !slices.Contains(retryOpts.codes, resp.StatusCode) {
					break
				}

				select {
				case <-timer.C:
					// calculate next delay
					timer.Reset(time.Duration(math.Pow(2, float64(i))) * retryOpts.baseDelay)
				}
			}

			if !timer.Stop() {
				<-timer.C
			}

			return resp, err
		})
	}
}
