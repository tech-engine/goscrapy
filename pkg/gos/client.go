package gos

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// round robin algo for proxy rotation
func RoundRobin(proxies ...string) func(*http.Request) (*url.URL, error) {
	proxyUrls := make([]*url.URL, 0, len(proxies))

	for _, proxy := range proxies {
		u, err := url.Parse(strings.TrimSpace(proxy))
		if err != nil {
			log.Panic(err)
			return nil
		}
		proxyUrls = append(proxyUrls, u)
	}

	var index uint32
	len := uint32(len(proxyUrls))
	return func(*http.Request) (*url.URL, error) {
		idx := atomic.AddUint32(&index, 1)
		// modulus is fine for now as it's not a bottleneck
		u := proxyUrls[(idx-1)%len]
		return u, nil
	}
}

// DefaultHTTPClient creates a default http client with defaults.
// If default values are set in the env it will pick the defaults from the env.
func DefaultHTTPClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()

	t.MaxIdleConns = MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN
	t.MaxConnsPerHost = MIDDLEWARE_DEFAULT_HTTP_MAX_CONN_PER_HOST
	t.MaxIdleConnsPerHost = MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN_PER_HOST

	if value, ok := os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN"); ok {
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			t.MaxIdleConns = int(v)
		}
	}

	if value, ok := os.LookupEnv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST"); ok {
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			t.MaxConnsPerHost = int(v)
		}
	}

	if value, ok := os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST"); ok {
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			t.MaxIdleConnsPerHost = int(v)
		}
	}

	if value, ok := os.LookupEnv("PROXY_LIST"); ok {
		proxies := strings.Split(value, ",")
		var cleanProxies []string
		for _, p := range proxies {
			if tp := strings.TrimSpace(p); tp != "" {
				cleanProxies = append(cleanProxies, tp)
			}
		}
		if len(cleanProxies) > 0 {
			t.Proxy = RoundRobin(cleanProxies...)
		}
	}

	cli := &http.Client{
		Timeout:   MIDDLEWARE_DEFAULT_HTTP_TIMEOUT_MS * time.Millisecond,
		Transport: t,
	}

	if value, ok := os.LookupEnv("MIDDLEWARE_HTTP_TIMEOUT_MS"); ok {
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			cli.Timeout = time.Duration(v) * time.Millisecond
		}
	}

	return cli
}
