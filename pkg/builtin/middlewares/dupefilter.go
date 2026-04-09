package middlewares

import (
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	"golang.org/x/crypto/blake2b"
)

var ERR_DUPEFILTER_BLOCKED = errors.New("duplicate request")

var hasherPool = sync.Pool{
	New: func() any {
		h, _ := blake2b.New256(nil)
		return h
	},
}

type RequestMap struct {
	seen map[[32]byte]struct{}
	mu   sync.RWMutex
}

func NewRequestMap() *RequestMap {
	return &RequestMap{
		seen: make(map[[32]byte]struct{}),
	}
}

func generateRequestFingerprint(r *http.Request) ([32]byte, error) {
	var (
		err  error
		body io.ReadCloser
	)

	if r.GetBody != nil {
		body, err = r.GetBody()
		if err != nil {
			return [32]byte{}, err
		}
		defer body.Close()
	}

	h := hasherPool.Get().(hash.Hash)
	defer hasherPool.Put(h)
	h.Reset()

	if body != nil {
		if _, err = io.Copy(h, body); err != nil {
			return [32]byte{}, err
		}
	}

	h.Write([]byte(r.Method))
	h.Write([]byte(r.URL.Host))
	h.Write([]byte(r.URL.Path))
	h.Write([]byte(r.URL.RawQuery))

	headerKeys := make([]string, 0, len(r.Header))
	for key := range r.Header {
		headerKeys = append(headerKeys, key)
	}

	sort.Strings(headerKeys)

	for _, key := range headerKeys {
		h.Write([]byte(key))
		for _, value := range r.Header[key] {
			h.Write([]byte(value))
		}
	}

	var finalHash [32]byte
	copy(finalHash[:], h.Sum(nil))

	return finalHash, nil
}

func DupeFilter(next http.RoundTripper) http.RoundTripper {
	requestMap := NewRequestMap()
	return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
		signature, err := generateRequestFingerprint(req)
		if err != nil {
			return nil, fmt.Errorf("dupefilter.go:DupeFilter: error generating request signature %w", err)
		}

		requestMap.mu.RLock()
		_, seen := requestMap.seen[signature]
		requestMap.mu.RUnlock()

		if seen {
			return nil, fmt.Errorf("dupefilter.go:DupeFilter: %w", ERR_DUPEFILTER_BLOCKED)
		}

		requestMap.mu.Lock()
		if _, ok := requestMap.seen[signature]; ok {
			requestMap.mu.Unlock()
			return nil, fmt.Errorf("dupefilter.go:DupeFilter: %w", ERR_DUPEFILTER_BLOCKED)
		}

		requestMap.seen[signature] = struct{}{}
		requestMap.mu.Unlock()

		return next.RoundTrip(req)
	})
}
