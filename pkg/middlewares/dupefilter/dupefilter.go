package dupefilter

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

var ERR_DUPEFILTER_BLOCKED = errors.New("duplicate request")

type RequestMap struct {
	seen map[string]bool
	mu   sync.RWMutex
}

func NewRequestMap() *RequestMap {
	return &RequestMap{
		seen: make(map[string]bool),
	}
}

func generateSHA1FingerprintFromReq(r *http.Request) (string, error) {

	var (
		err error
	)

	bodyBuf := new(bytes.Buffer)
	combinedBuf := new(bytes.Buffer)

	hash := sha1.New()

	if r.Body != nil {
		teeReader := io.TeeReader(r.Body, bodyBuf)

		if _, err = io.Copy(hash, teeReader); err != nil {
			return "", err
		}
	}

	combinedBuf.WriteString(r.Method)
	combinedBuf.WriteString(r.URL.String())

	headerKeys := make([]string, 0, len(r.Header))
	for key := range r.Header {
		headerKeys = append(headerKeys, key)
	}

	sort.Strings(headerKeys)

	// added sorted headers
	for _, key := range headerKeys {
		values := r.Header[key]
		for _, value := range values {
			combinedBuf.WriteString(key)
			combinedBuf.WriteString(value)
		}
	}

	if _, err = combinedBuf.Write(hash.Sum(nil)); err != nil {
		return "", err
	}

	finalHash := sha1.Sum(combinedBuf.Bytes())
	r.Body = io.NopCloser(bodyBuf)

	return hex.EncodeToString(finalHash[:]), nil

}

func DupeFilterMiddleware(next http.RoundTripper) http.RoundTripper {
	requestMap := NewRequestMap()
	return core.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
		signature, err := generateSHA1FingerprintFromReq(req)

		if err != nil {
			return nil, fmt.Errorf("DupeFilterMiddleware: error generating request signature %w", err)
		}
		
		requestMap.mu.Lock()

		// we have already seen this signature so we skip
		if _, ok := requestMap.seen[signature]; ok {
			requestMap.mu.Unlock()
			return nil, fmt.Errorf("DupeFilterMiddleware: %w", ERR_DUPEFILTER_BLOCKED)
		}

		requestMap.seen[signature] = true
		requestMap.mu.Unlock()

		return next.RoundTrip(req)
	})
}
