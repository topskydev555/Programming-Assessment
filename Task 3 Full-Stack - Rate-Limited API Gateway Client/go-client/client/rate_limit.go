package client

import (
	"fmt"
	"net/http"
	"time"
)

type RateLimiter struct {
	tokens chan struct{}
}

func NewRateLimiter(requestsPerSecond int, burst int) (*RateLimiter, error) {
	if requestsPerSecond <= 0 {
		return nil, fmt.Errorf("requestsPerSecond must be > 0")
	}
	if burst <= 0 {
		return nil, fmt.Errorf("burst must be > 0")
	}

	limiter := &RateLimiter{
		tokens: make(chan struct{}, burst),
	}

	for i := 0; i < burst; i++ {
		limiter.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(requestsPerSecond)
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			select {
			case limiter.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return limiter, nil
}

func WithRateLimit(limiter *RateLimiter) Middleware {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if req == nil {
				return nil, fmt.Errorf("request cannot be nil")
			}
			if limiter == nil {
				return next.Do(req)
			}
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-limiter.tokens:
				return next.Do(req)
			}
		})
	}
}
