package middlewaremanager

import "net/http"

type Middleware func(next http.RoundTripper) http.RoundTripper

type MiddlewareFunc func(req *http.Request) (*http.Response, error)

func (mf MiddlewareFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return mf(req)
}

type MiddlewareManager struct {
	httpClient *http.Client
}

func New(cli *http.Client) *MiddlewareManager {
	return &MiddlewareManager{
		httpClient: cli,
	}
}

func (m *MiddlewareManager) HTTPClient() *http.Client {
	return m.httpClient
}

func (m *MiddlewareManager) Add(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		m.httpClient.Transport = middleware(m.httpClient.Transport)
	}
}
