package httpclient

import "net/http"

//go:generate counterfeiter -o httpclientfakes/fake_client.go . Client
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
