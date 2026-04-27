package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type stubLogger struct {
	mu      sync.Mutex
	entries []string
}

func (l *stubLogger) Logf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, format)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRetryLayer_Isolated(t *testing.T) {
	attempts := 0
	next := DoerFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return response(500, "fail"), nil
		}
		return response(200, "ok"), nil
	})

	client := Compose(next, WithRetry(RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
	}))

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/items", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestCacheLayer_Isolated(t *testing.T) {
	calls := 0
	next := DoerFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return response(200, "cached"), nil
	})

	cache := NewMemoryCache(time.Minute)
	client := Compose(next, WithResponseCache(cache))

	req1, _ := http.NewRequest(http.MethodGet, "https://example.com/user", nil)
	if _, err := client.Do(req1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req2, _ := http.NewRequest(http.MethodGet, "https://example.com/user", nil)
	if _, err := client.Do(req2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected underlying client call once, got %d", calls)
	}
}

func TestRateLimitLayer_Isolated(t *testing.T) {
	limiter, err := NewRateLimiter(1, 1)
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	calls := 0
	next := DoerFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return response(200, "ok"), nil
	})
	client := Compose(next, WithRateLimit(limiter))

	req1, _ := http.NewRequest(http.MethodGet, "https://example.com/a", nil)
	if _, err := client.Do(req1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/b", nil)
	_, err = client.Do(req2)
	if err == nil {
		t.Fatalf("expected context timeout while waiting for token")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error type: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one successful call, got %d", calls)
	}
}

func TestLoggingLayer_Isolated(t *testing.T) {
	logger := &stubLogger{}
	next := DoerFunc(func(req *http.Request) (*http.Response, error) {
		return response(200, "ok"), nil
	})

	client := Compose(next, WithLogging(logger))
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/log", nil)
	if _, err := client.Do(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logger.entries))
	}
}

func TestComposedLayers_WorkTogether(t *testing.T) {
	attempts := 0
	logger := &stubLogger{}
	cache := NewMemoryCache(time.Minute)
	limiter, err := NewRateLimiter(20, 2)
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	base := DoerFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return response(500, "transient"), nil
		}
		return response(200, "payload"), nil
	})

	client := Compose(
		base,
		WithLogging(logger),
		WithRateLimit(limiter),
		WithRetry(RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond}),
		WithResponseCache(cache),
	)

	req1, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.StatusCode != 200 {
		t.Fatalf("unexpected first response status: %d", resp1.StatusCode)
	}

	req2, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.StatusCode != 200 {
		t.Fatalf("unexpected second response status: %d", resp2.StatusCode)
	}

	if attempts != 2 {
		t.Fatalf("expected exactly 2 base attempts (retry then cache hit), got %d", attempts)
	}
	if len(logger.entries) != 2 {
		t.Fatalf("expected two logs, got %d", len(logger.entries))
	}
}
