package mutlicookiejar

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/middleware"
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

// CookieJar returns a CookieJar corresponding to a key or create one if key doesn't exist
func (m *MultiCookieJar) CookieJar(key string) http.CookieJar {

	m.mu.RLock()
	defer m.mu.RUnlock()

	jar, ok := m.jars[key]

	if !ok {
		return nil
	}

	return jar
}

// EncodeCookieJar returns coo
func (m *MultiCookieJar) EncodeCookieJar(key string, _url string) string {

	m.mu.RLock()
	defer m.mu.RUnlock()

	jar, ok := m.jars[key]

	if !ok {
		return ""
	}

	var buf bytes.Buffer
	__url, _ := url.Parse(_url)

	for _, cookie := range jar.Cookies(__url) {
		buf.WriteString(fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value))
	}

	return buf.String()

}

func (m *MultiCookieJar) SetCookieJar(key string, jar http.CookieJar) http.CookieJar {

	m.mu.Lock()
	defer m.mu.Unlock()

	key = strings.Trim(key, " ")

	if jar == nil {
		jar, _ = cookiejar.New(nil)
	}

	m.jars[key] = jar
	return jar
}

func MultiCookieJarMiddleware(next http.RoundTripper) http.RoundTripper {
	mCookieJar := NewMultiCookieJar()
	return middleware.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
		cookieJarKey := strings.Trim(req.Header.Get("cookie-jar"), " ")

		// try picking cookies from jar corresponding to a key
		jar := mCookieJar.CookieJar(cookieJarKey)

		if jar == nil {
			// create an empty jar
			jar = mCookieJar.SetCookieJar(cookieJarKey, nil)
		}

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
