package client

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type cacheEntry struct {
	expiresAt  time.Time
	statusCode int
	headers    http.Header
	body       []byte
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
	now   func() time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	if ttl <= 0 {
		ttl = time.Second
	}
	return &MemoryCache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
		now:   time.Now,
	}
}

func cacheKey(req *http.Request) string {
	return req.Method + ":" + req.URL.String()
}

func WithResponseCache(cache *MemoryCache) Middleware {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if req == nil {
				return nil, fmt.Errorf("request cannot be nil")
			}
			if cache == nil || req.Method != http.MethodGet {
				return next.Do(req)
			}

			key := cacheKey(req)
			now := cache.now()

			cache.mu.RLock()
			entry, exists := cache.items[key]
			cache.mu.RUnlock()

			if exists && now.Before(entry.expiresAt) {
				return buildResponseFromCache(entry.statusCode, entry.headers, entry.body, req), nil
			}

			resp, err := next.Do(req)
			if err != nil || resp == nil || resp.StatusCode >= 400 {
				return resp, err
			}

			bodyBytes, err := cloneResponse(resp)
			if err != nil {
				return nil, err
			}

			cache.mu.Lock()
			cache.items[key] = cacheEntry{
				expiresAt:  now.Add(cache.ttl),
				statusCode: resp.StatusCode,
				headers:    resp.Header.Clone(),
				body:       bodyBytes,
			}
			cache.mu.Unlock()

			return resp, nil
		})
	}
}
