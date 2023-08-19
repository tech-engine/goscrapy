package core

import "net/http"

type Middleware func(next http.RoundTripper) http.RoundTripper

type MiddlewareFunc func(req *http.Request) (*http.Response, error)

func (mf MiddlewareFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return mf(req)
}
