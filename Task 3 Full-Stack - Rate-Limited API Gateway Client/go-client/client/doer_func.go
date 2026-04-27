package client

import "net/http"

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DoerFunc func(req *http.Request) (*http.Response, error)

func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
