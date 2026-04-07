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

type Browser string

const (
	BrowserChrome  Browser = "chrome"
	BrowserFirefox Browser = "firefox"
	BrowserSafari  Browser = "safari"
	BrowserEdge    Browser = "edge"
)

type AzureTLSOptions struct {
	Browser    Browser
	Proxy      string
	SessionKey string
	JA3        string
}

type azureTLSCtxKey struct{}

func WithAzureTLSOptions(ctx context.Context, opts *AzureTLSOptions) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
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
func (a *azureTLS) getSession(opts *AzureTLSOptions) (*azuretls.Session, error) {
	a.mu.RLock()
	sessionKey := opts.SessionKey
	if sessionKey == "" {
		sessionKey = "default"
	}

	if session, exists := a.sessions[sessionKey]; exists {
		a.mu.RUnlock()
		return session, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	if session, exists := a.sessions[sessionKey]; exists {
		return session, nil
	}

	session := azuretls.NewSession()

	switch opts.Browser {
	case BrowserFirefox:
		session.Browser = azuretls.Firefox
	case BrowserSafari:
		session.Browser = azuretls.Safari
	case BrowserEdge:
		session.Browser = azuretls.Edge
	default:
		session.Browser = azuretls.Chrome
	}

	if opts.Proxy != "" {
		session.SetProxy(opts.Proxy)
	}

	if opts.JA3 != "" {
		if err := session.ApplyJa3(opts.JA3, session.Browser); err != nil {
			return nil, fmt.Errorf("failed to apply JA3 fingerprint: %w", err)
		}
	}

	a.sessions[sessionKey] = session
	return session, nil
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

		if ctxOpts.JA3 != "" {
			mergedOpts.JA3 = ctxOpts.JA3
		}

		opts = &mergedOpts
	}

	tlsSession, err := a.getSession(opts)
	if err != nil {
		return nil, err
	}

	// map http.Header to fhttp.Header
	azureHeaders := make(fhttp.Header, len(req.Header))
	fmt.Println("my headers", req.Header)
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

	httpHeader := make(http.Header, len(azureResp.HttpResponse.Header))
	maps.Copy(httpHeader, azureResp.HttpResponse.Header)

	// map azuretls.Response to http.Response
	resp := &http.Response{
		Status:        azureResp.HttpResponse.Status,
		StatusCode:    azureResp.HttpResponse.StatusCode,
		Proto:         azureResp.HttpResponse.Proto,
		ProtoMajor:    azureResp.HttpResponse.ProtoMajor,
		ProtoMinor:    azureResp.HttpResponse.ProtoMinor,
		Header:        httpHeader,
		Body:          io.NopCloser(bytes.NewReader(azureResp.Body)),
		ContentLength: azureResp.HttpResponse.ContentLength,
		Uncompressed:  azureResp.HttpResponse.Uncompressed,
		Request:       req,
	}
	return resp, nil
}

func AzureTLS(globalOpts *AzureTLSOptions) func(http.RoundTripper) http.RoundTripper {
	if globalOpts == nil {
		globalOpts = &AzureTLSOptions{Browser: BrowserChrome}
	}
	az := newAzureTLS(globalOpts)
	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			return az.roundTrip(req)
		})
	}
}
