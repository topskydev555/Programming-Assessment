package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	ShouldRetry func(resp *http.Response, err error) bool
}

func DefaultShouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp != nil && resp.StatusCode >= 500
}

func WithRetry(cfg RetryConfig) Middleware {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 50 * time.Millisecond
	}
	if cfg.ShouldRetry == nil {
		cfg.ShouldRetry = DefaultShouldRetry
	}

	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if req == nil {
				return nil, fmt.Errorf("request cannot be nil")
			}

			bodyBytes, err := copyRequestBody(req)
			if err != nil {
				return nil, err
			}

			var lastResp *http.Response
			var lastErr error

			for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
				attemptReq := req.Clone(req.Context())
				if bodyBytes != nil {
					attemptReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}

				lastResp, lastErr = next.Do(attemptReq)
				if !cfg.ShouldRetry(lastResp, lastErr) || attempt == cfg.MaxAttempts {
					return lastResp, lastErr
				}

				delay := cfg.BaseDelay * time.Duration(1<<(attempt-1))
				timer := time.NewTimer(delay)
				select {
				case <-req.Context().Done():
					timer.Stop()
					return nil, req.Context().Err()
				case <-timer.C:
				}
			}

			return lastResp, lastErr
		})
	}
}
