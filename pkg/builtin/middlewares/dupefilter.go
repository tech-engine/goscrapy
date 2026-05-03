package middlewares

import (
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"slices"
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

// Represents the interface for deduplication backends.
type IDupeStore interface {
	// must returns true if already seen, otherwise false.
	Add(key [32]byte) bool
}

type Config struct {
	Store      IDupeStore
	MaxEntries int
}

// default inmemory deduplication store
type mapStore struct {
	seen       map[[32]byte]struct{}
	mu         sync.RWMutex
	maxEntries int
}

func newMapStore(maxEntries int) *mapStore {
	initialCap := 1024
	if maxEntries > 0 && maxEntries < initialCap {
		initialCap = maxEntries
	}
	return &mapStore{
		seen:       make(map[[32]byte]struct{}, initialCap),
		maxEntries: maxEntries,
	}
}

func (s *mapStore) Add(key [32]byte) bool {
	s.mu.RLock()
	_, seen := s.seen[key]
	s.mu.RUnlock()
	if seen {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.seen[key]; ok {
		return true
	}

	// stop tracking if limit reached, if maxEntries=0, it' unlimited, please note
	if s.maxEntries > 0 && len(s.seen) >= s.maxEntries {
		return false
	}

	s.seen[key] = struct{}{}
	return false
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

	slices.Sort(headerKeys)

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

func DupeFilter(cfg ...*Config) middlewaremanager.Middleware {
	var c *Config
	if len(cfg) > 0 && cfg[0] != nil {
		c = cfg[0]
	} else {
		c = &Config{}
	}

	var store IDupeStore
	if c.Store != nil {
		store = c.Store
	} else {
		store = newMapStore(c.MaxEntries)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			signature, err := generateRequestFingerprint(req)
			if err != nil {
				return nil, fmt.Errorf("dupefilter.go:DupeFilter: error generating request signature %w", err)
			}

			if store.Add(signature) {
				return nil, fmt.Errorf("dupefilter.go:DupeFilter: %w", ERR_DUPEFILTER_BLOCKED)
			}

			return next.RoundTrip(req)
		})
	}
}
