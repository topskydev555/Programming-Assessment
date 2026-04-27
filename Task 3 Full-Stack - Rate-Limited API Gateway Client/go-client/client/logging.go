package client

import (
	"fmt"
	"net/http"
	"time"
)

func WithLogging(logger Logger) Middleware {
	return func(next Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			if req == nil {
				return nil, fmt.Errorf("request cannot be nil")
			}

			start := time.Now()
			resp, err := next.Do(req)
			duration := time.Since(start)

			if logger != nil {
				statusCode := 0
				if resp != nil {
					statusCode = resp.StatusCode
				}
				logger.Logf("method=%s url=%s status=%d duration=%s err=%v", req.Method, req.URL.String(), statusCode, duration, err)
			}

			return resp, err
		})
	}
}
