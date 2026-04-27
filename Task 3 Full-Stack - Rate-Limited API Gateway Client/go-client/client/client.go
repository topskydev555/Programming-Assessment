package client

import "net/http"

type Layer func(next Doer) Doer

func Chain(base Doer, layers ...Layer) Doer {
	wrapped := base
	for i := len(layers) - 1; i >= 0; i-- {
		wrapped = layers[i](wrapped)
	}
	return wrapped
}

type Client struct {
	doer Doer
}

func New(base Doer, layers ...Layer) *Client {
	return &Client{doer: Chain(base, layers...)}
}

func NewHTTPClient(httpClient *http.Client, layers ...Layer) *Client {
	return New(httpClient, layers...)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.doer.Do(req)
}
