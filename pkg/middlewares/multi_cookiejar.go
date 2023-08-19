package middlewares

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type MultiCookieJar struct {
	jars map[string]http.CookieJar
	mu   sync.RWMutex
}

// NewMultiCookieJar creates a new MultiCookieJar.
func NewMultiCookieJar() *MultiCookieJar {
	return &MultiCookieJar{
		jars: make(map[string]http.CookieJar),
	}
}

func (m *MultiCookieJar) CookieJar(key string) http.CookieJar {

	m.mu.Lock()
	defer m.mu.Unlock()

	jar, ok := m.jars[key]
	if ok {
		return jar
	}

	jar, _ = cookiejar.New(nil)

	m.jars[key] = jar
	return jar
}

func MultiCookieJarMiddleware(next http.RoundTripper) http.RoundTripper {
	mCookieJar := NewMultiCookieJar()
	return core.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
		cookieJarKey := strings.Trim(req.Header.Get("cookie-jar"), " ")

		// pick cookies from jar and add cookies to request
		jar := mCookieJar.CookieJar(cookieJarKey)
		reqCookies := jar.Cookies(req.URL)

		for _, rc := range reqCookies {
			req.AddCookie(rc)
		}

		// remove cookie-jar header
		req.Header.Del("cookie-jar")

		resp, err := next.RoundTrip(req)

		// update cookies
		jar.SetCookies(req.URL, resp.Cookies())

		return resp, err
	})
}
