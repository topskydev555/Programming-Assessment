package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Logger interface {
	Logf(format string, args ...any)
}

type BackoffFunc func(attempt int) <-chan timeSignal

type timeSignal <-chan struct{}

type Middleware func(next Doer) Doer

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPClient{client: httpClient}
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func Compose(base Doer, middlewares ...Middleware) Doer {
	wrapped := base
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

func cloneResponse(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes, nil
}

func buildResponseFromCache(statusCode int, headers http.Header, body []byte, req *http.Request) *http.Response {
	headerCopy := make(http.Header, len(headers))
	for k, v := range headers {
		copied := make([]string, len(v))
		copy(copied, v)
		headerCopy[k] = copied
	}

	return &http.Response{
		StatusCode: statusCode,
		Header:     headerCopy,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}
}

func copyRequestBody(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil {
		return nil, nil
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes, nil
}

func cloneRequestWithBody(req *http.Request, body []byte) *http.Request {
	cloned := req.Clone(req.Context())
	if body != nil {
		cloned.Body = io.NopCloser(bytes.NewReader(body))
	}
	return cloned
}

func waitWithContext(ctx context.Context, sig <-chan struct{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sig:
		return nil
	}
}
