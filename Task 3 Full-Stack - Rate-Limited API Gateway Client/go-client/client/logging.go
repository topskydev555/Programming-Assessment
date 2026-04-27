package client

import (
	"net/http"
	"time"
)

type Logger interface {
	LogRequest(req *http.Request)
	LogResponse(req *http.Request, resp *http.Response, err error, duration time.Duration)
}

func WithLogging(logger Logger) Layer {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			logger.LogRequest(req)
			resp, err := next.Do(req)
			logger.LogResponse(req, resp, err, time.Since(start))
			return resp, err
		})
	}
}
