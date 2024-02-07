package middlewares

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

type multiCookieJar struct {
	jars map[string]http.CookieJar
	mu   sync.RWMutex
}

// NewMultiCookieJar creates a new MultiCookieJar.
func NewMultiCookieJar() *multiCookieJar {
	return &multiCookieJar{
		jars: make(map[string]http.CookieJar),
	}
}

// GetCookieJar returns a CookieJar corresponding to a key or create one if key doesn't exist
func (m *multiCookieJar) GetCookieJar(key string) http.CookieJar {

	m.mu.Lock()
	defer m.mu.Unlock()
	jar, ok := m.jars[key]

	// in case we don't have a cookie jar based on the key, we create a new one
	if !ok {
		jar, _ = cookiejar.New(nil)
	}
	return jar
}

func MultiCookieJar(next http.RoundTripper) http.RoundTripper {
	mCookieJar := NewMultiCookieJar()
	return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
		cookieJarKey := strings.Trim(req.Header.Get("COOKIE_JAR_KEY"), " ")

		jar := mCookieJar.GetCookieJar(cookieJarKey)

		reqCookies := jar.Cookies(req.URL)

		for _, rc := range reqCookies {
			req.AddCookie(rc)
		}

		// remove cookie_jar_key header
		req.Header.Del("COOKIE_JAR_KEY")

		resp, err := next.RoundTrip(req)

		// update cookies
		jar.SetCookies(req.URL, resp.Cookies())

		return resp, err
	})
}
