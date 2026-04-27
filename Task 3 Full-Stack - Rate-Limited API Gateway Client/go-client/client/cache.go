package client

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"
)

type cacheEntry struct {
	statusCode int
	header     http.Header
	body       []byte
	expiresAt  time.Time
}

type TTLCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
	nowFn   func() time.Time
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		nowFn:   time.Now,
	}
}

func cacheKey(req *http.Request) string {
	return req.Method + ":" + req.URL.String()
}

func WithResponseCache(cache *TTLCache) Layer {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet || cache.ttl <= 0 {
				return next.Do(req)
			}

			key := cacheKey(req)
			now := cache.nowFn()
			if resp := cache.get(key, now); resp != nil {
				return resp, nil
			}

			resp, err := next.Do(req)
			if err != nil || resp == nil {
				return resp, err
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return resp, nil
			}

			cachedResp, body, readErr := cloneResponse(resp)
			if readErr != nil {
				return resp, nil
			}

			cache.set(key, cacheEntry{
				statusCode: cachedResp.StatusCode,
				header:     cachedResp.Header.Clone(),
				body:       body,
				expiresAt:  now.Add(cache.ttl),
			})

			return cachedResp, nil
		})
	}
}

func (c *TTLCache) get(key string, now time.Time) *http.Response {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	if now.After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil
	}
	return entry.toResponse()
}

func (c *TTLCache) set(key string, entry cacheEntry) {
	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
}

func (e cacheEntry) toResponse() *http.Response {
	return &http.Response{
		StatusCode: e.statusCode,
		Header:     e.header.Clone(),
		Body:       io.NopCloser(bytes.NewReader(e.body)),
	}
}

func cloneResponse(resp *http.Response) (*http.Response, []byte, error) {
	if resp.Body == nil {
		empty := []byte{}
		resp.Body = io.NopCloser(bytes.NewReader(empty))
		return &http.Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       io.NopCloser(bytes.NewReader(empty)),
		}, empty, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))

	cloned := &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}

	return cloned, body, nil
}
