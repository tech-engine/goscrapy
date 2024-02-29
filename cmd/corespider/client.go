package corespider

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
)

type clientOpts struct {
	timeout time.Duration
	transportOpts
}

type transportOpts struct {
	proxyFn                                            func(*http.Request) (*url.URL, error)
	maxIdleConns, maxConnsPerHost, maxIdleConnsPerHost int
}

func defaultClientOpts() clientOpts {
	opts := clientOpts{
		timeout: MIDDLEWARE_DEFAULT_HTTP_TIMEOUT_MS * time.Millisecond,
		transportOpts: transportOpts{
			proxyFn:             nil,
			maxIdleConns:        MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN,
			maxConnsPerHost:     MIDDLEWARE_DEFAULT_HTTP_MAX_CONN_PER_HOST,
			maxIdleConnsPerHost: MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN_PER_HOST,
		},
	}

	value, ok := os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN")

	if ok {
		maxIdleConn, err := strconv.Atoi(value)
		if err == nil {
			opts.maxIdleConns = maxIdleConn
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST")

	if ok {
		maxConnPerHost, err := strconv.Atoi(value)
		if err == nil {
			opts.maxConnsPerHost = maxConnPerHost
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST")

	if ok {
		maxIdleConnPerHost, err := strconv.Atoi(value)
		if err == nil {
			opts.maxConnsPerHost = maxIdleConnPerHost
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_DEFAULT_HTTP_TIMEOUT_MS")

	if ok {
		timeoutMs, err := strconv.Atoi(value)
		if err == nil {
			opts.timeout = time.Duration(timeoutMs) * time.Millisecond
		}
	}

	return opts
}

func WithTimeout(t time.Duration) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		opts.timeout = t
	}
}

func WithMaxIdleConns(maxIdleConns int) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		opts.maxIdleConns = maxIdleConns
	}
}

func WithMaxConnsPerHost(maxConnsPerHost int) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		opts.maxConnsPerHost = maxConnsPerHost
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		opts.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithProxyFn(fn func(*http.Request) (*url.URL, error)) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		opts.proxyFn = fn
	}
}

func WithProxies(proxies ...string) types.OptFunc[clientOpts] {
	return func(opts *clientOpts) {
		proxyUrls := make([]*url.URL, 0, len(proxies))

		for _, proxy := range proxies {
			u, err := url.Parse(strings.TrimSpace(proxy))
			if err != nil {
				log.Panic(err)
				return
			}
			proxyUrls = append(proxyUrls, u)
		}
		opts.proxyFn = roundRobin(proxyUrls)
	}
}

// round robin algo for proxy rotation
func roundRobin(urls []*url.URL) func(*http.Request) (*url.URL, error) {
	var index uint32
	len := uint32(len(urls))
	return func(*http.Request) (*url.URL, error) {
		index := atomic.AddUint32(&index, 1)
		u := urls[(index-1)%len]
		return u, nil
	}
}

// createDefaultHTTPClient creates a default http client with defaults.
// If default values are set in the env it will pick the defaults from the env.
func DefaultClient(opts ...types.OptFunc[clientOpts]) *http.Client {
	cli := &http.Client{}

	// load in default options
	cliOpts := defaultClientOpts()

	for _, opt := range opts {
		opt(&cliOpts)
	}

	t := http.DefaultTransport.(*http.Transport).Clone()

	// set all value from transport options
	t.MaxIdleConns = cliOpts.maxIdleConns
	t.MaxConnsPerHost = cliOpts.maxConnsPerHost
	t.MaxIdleConnsPerHost = cliOpts.maxIdleConnsPerHost
	t.Proxy = cliOpts.proxyFn

	// set client options
	cli.Timeout = cliOpts.timeout

	cli.Transport = t

	return cli
}
