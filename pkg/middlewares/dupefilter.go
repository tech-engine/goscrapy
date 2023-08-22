package middlewares

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"sort"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

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
		bodyBuf, combinedBuf *bytes.Buffer
		err                  error
	)

	teeReader := io.TeeReader(r.Body, bodyBuf)

	hash := sha1.New()

	if _, err = io.Copy(hash, teeReader); err != nil {
		return "", err
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
			return nil, err
		}

		// we have already seen this signature so we skip
		if _, ok := requestMap.seen[signature]; ok {
			return nil, nil
		}

		requestMap.seen[signature] = true

		return next.RoundTrip(req)
	})
}
