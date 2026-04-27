package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeLimiter struct {
	calls int
	err   error
}

func (f *fakeLimiter) Wait(_ context.Context) error {
	f.calls++
	return f.err
}

type fakeSleeper struct {
	calls []time.Duration
}

func (f *fakeSleeper) Sleep(d time.Duration) {
	f.calls = append(f.calls, d)
}

type fakeLogger struct {
	requests  int
	responses int
}

func (f *fakeLogger) LogRequest(_ *http.Request) {
	f.requests++
}

func (f *fakeLogger) LogResponse(_ *http.Request, _ *http.Response, _ error, _ time.Duration) {
	f.responses++
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestWithRateLimit_WaitsBeforeRequest(t *testing.T) {
	limiter := &fakeLimiter{}
	called := 0
	layer := WithRateLimit(limiter)
	doer := layer(DoerFunc(func(_ *http.Request) (*http.Response, error) {
		called++
		return response(200, "ok"), nil
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := doer.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limiter.calls != 1 {
		t.Fatalf("expected limiter to be called once, got %d", limiter.calls)
	}
	if called != 1 {
		t.Fatalf("expected downstream to be called once, got %d", called)
	}
}

func TestWithRetry_RetriesServerErrors(t *testing.T) {
	sleeper := &fakeSleeper{}
	attempts := 0
	layer := WithRetry(RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		Sleeper:     sleeper,
		ShouldRetry: func(resp *http.Response, err error) bool {
			return err != nil || resp.StatusCode >= 500
		},
	})

	doer := layer(DoerFunc(func(_ *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return response(500, "fail"), nil
		}
		return response(200, "ok"), nil
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := doer.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if len(sleeper.calls) != 2 {
		t.Fatalf("expected 2 backoffs, got %d", len(sleeper.calls))
	}
}

func TestWithResponseCache_CachesGetWithinTTL(t *testing.T) {
	cache := NewTTLCache(10 * time.Second)
	calls := 0
	layer := WithResponseCache(cache)
	doer := layer(DoerFunc(func(_ *http.Request) (*http.Response, error) {
		calls++
		return response(200, "cached"), nil
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com/items", nil)
	_, _ = doer.Do(req)
	resp, err := doer.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected downstream call once, got %d", calls)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "cached" {
		t.Fatalf("expected cached body, got %q", string(body))
	}
}

func TestWithLogging_LogsBothRequestAndResponse(t *testing.T) {
	logger := &fakeLogger{}
	layer := WithLogging(logger)
	doer := layer(DoerFunc(func(_ *http.Request) (*http.Response, error) {
		return response(200, "ok"), nil
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := doer.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger.requests != 1 || logger.responses != 1 {
		t.Fatalf("expected one request and one response log, got %d/%d", logger.requests, logger.responses)
	}
}

func TestComposedLayers_WorkTogether(t *testing.T) {
	limiter := &fakeLimiter{}
	sleeper := &fakeSleeper{}
	logger := &fakeLogger{}
	cache := NewTTLCache(10 * time.Second)
	attempts := 0

	base := DoerFunc(func(_ *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return response(500, "retry"), nil
		}
		return response(200, "ok"), nil
	})

	client := New(base,
		WithLogging(logger),
		WithRateLimit(limiter),
		WithRetry(RetryPolicy{
			MaxAttempts: 2,
			BaseDelay:   time.Millisecond,
			Sleeper:     sleeper,
			ShouldRetry: DefaultRetryPolicy().ShouldRetry,
		}),
		WithResponseCache(cache),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/composed", nil)
	resp1, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp1.StatusCode)
	}

	resp2, err := client.Do(httptest.NewRequest(http.MethodGet, "http://example.com/composed", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp2.StatusCode)
	}

	if attempts != 2 {
		t.Fatalf("expected base doer attempts 2 (first call retries, second cached), got %d", attempts)
	}
	if limiter.calls != 2 {
		t.Fatalf("expected limiter to run once per top-level request, got %d", limiter.calls)
	}
	if logger.requests != 2 || logger.responses != 2 {
		t.Fatalf("expected two log calls, got %d/%d", logger.requests, logger.responses)
	}
}
