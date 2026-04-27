package client

import (
	"errors"
	"net/http"
	"time"
)

type Sleeper interface {
	Sleep(d time.Duration)
}

type realSleeper struct{}

func (realSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	ShouldRetry func(resp *http.Response, err error) bool
	Sleeper     Sleeper
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		ShouldRetry: func(resp *http.Response, err error) bool {
			if err != nil {
				return true
			}
			return resp != nil && resp.StatusCode >= 500
		},
		Sleeper: realSleeper{},
	}
}

func WithRetry(policy RetryPolicy) Layer {
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	if policy.BaseDelay < 0 {
		policy.BaseDelay = 0
	}
	if policy.ShouldRetry == nil {
		policy.ShouldRetry = DefaultRetryPolicy().ShouldRetry
	}
	if policy.Sleeper == nil {
		policy.Sleeper = realSleeper{}
	}

	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			var lastResp *http.Response
			var lastErr error

			for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
				resp, err := next.Do(req)
				lastResp = resp
				lastErr = err

				if !policy.ShouldRetry(resp, err) || attempt == policy.MaxAttempts {
					return resp, err
				}

				backoff := policy.BaseDelay * time.Duration(1<<(attempt-1))
				policy.Sleeper.Sleep(backoff)
			}

			return lastResp, lastErr
		})
	}
}

var ErrRetryExhausted = errors.New("retry exhausted")
