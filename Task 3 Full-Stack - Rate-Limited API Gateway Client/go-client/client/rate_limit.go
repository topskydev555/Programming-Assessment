package client

import (
	"context"
	"net/http"
)

type Limiter interface {
	Wait(ctx context.Context) error
}

func WithRateLimit(limiter Limiter) Layer {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if err := limiter.Wait(req.Context()); err != nil {
				return nil, err
			}
			return next.Do(req)
		})
	}
}
