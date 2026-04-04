package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"sync"

	"github.com/Noooste/azuretls-client"
	fhttp "github.com/Noooste/fhttp"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

type AzureTLSOptions struct {
	Browser    string
	Proxy      string
	SessionKey string
}

type azureTLSCtxKey struct{}

func WithAzureTLSOptions(ctx context.Context, opts *AzureTLSOptions) context.Context {
	return context.WithValue(ctx, azureTLSCtxKey{}, opts)
}

type azureTLS struct {
	globalOpts *AzureTLSOptions
	sessions   map[string]*azuretls.Session
	mu         sync.RWMutex
}

func newAzureTLS(globalOpts *AzureTLSOptions) *azureTLS {
	return &azureTLS{
		globalOpts: globalOpts,
		sessions:   make(map[string]*azuretls.Session),
	}
}

// getSession retrieves or provisions a session safely
func (a *azureTLS) getSession(opts *AzureTLSOptions) *azuretls.Session {
	a.mu.RLock()
	sessionKey := opts.SessionKey
	if sessionKey == "" {
		sessionKey = "default"
	}

	if session, exists := a.sessions[sessionKey]; exists {
		a.mu.RUnlock()
		return session
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if session, exists := a.sessions[sessionKey]; exists {
		return session
	}

	session := azuretls.NewSession()

	switch opts.Browser {
	case "firefox":
		session.Browser = azuretls.Firefox
	default:
		session.Browser = azuretls.Chrome
	}

	if opts.Proxy != "" {
		session.SetProxy(opts.Proxy)
	}

	a.sessions[sessionKey] = session
	return session
}

func (a *azureTLS) roundTrip(req *http.Request) (*http.Response, error) {
	opts := a.globalOpts

	if ctxOpts, ok := req.Context().Value(azureTLSCtxKey{}).(*AzureTLSOptions); ok {
		mergedOpts := *opts

		if ctxOpts.Browser != "" {
			mergedOpts.Browser = ctxOpts.Browser
		}
		if ctxOpts.Proxy != "" {
			mergedOpts.Proxy = ctxOpts.Proxy
		}
		if ctxOpts.SessionKey != "" {
			mergedOpts.SessionKey = ctxOpts.SessionKey
		}

		opts = &mergedOpts
	}

	tlsSession := a.getSession(opts)

	// map http.Header to fhttp.Header
	azureHeaders := make(fhttp.Header, len(req.Header))

	maps.Copy(azureHeaders, req.Header)

	azureResp, err := tlsSession.Do(&azuretls.Request{
		Method: req.Method,
		Url:    req.URL.String(),
		Header: azureHeaders,
		Body:   req.Body,
		// IgnoreBody: true,
	})

	if err != nil {
		return nil, fmt.Errorf("AzureTLS failed: %w", err)
	}

	stdHeader := make(http.Header, len(azureResp.HttpResponse.Header))
	maps.Copy(stdHeader, azureResp.HttpResponse.Header)

	// map azuretls.Response to http.Response
	resp := &http.Response{
		Status:        azureResp.HttpResponse.Status,
		StatusCode:    azureResp.HttpResponse.StatusCode,
		Proto:         azureResp.HttpResponse.Proto,
		ProtoMajor:    azureResp.HttpResponse.ProtoMajor,
		ProtoMinor:    azureResp.HttpResponse.ProtoMinor,
		Header:        stdHeader,
		Body:          io.NopCloser(bytes.NewReader(azureResp.Body)),
		ContentLength: azureResp.HttpResponse.ContentLength,
		Uncompressed:  azureResp.HttpResponse.Uncompressed,
		Request:       req,
	}
	return resp, nil
}

func AzureTLS(globalOpts *AzureTLSOptions) func(http.RoundTripper) http.RoundTripper {
	az := newAzureTLS(globalOpts)
	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			return az.roundTrip(req)
		})
	}
}
