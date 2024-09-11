package httpclient

import (
	"net/http"
)

// Client is the global HTTP client
var Client *http.Client

// InitClient initializes the global HTTP client with an Authorization header.
func InitClient(authToken string) {
	Client = &http.Client{
		Transport: &transportWithAuth{
			authToken: authToken,
			rt:        http.DefaultTransport, // Use default transport for requests
		},
	}
}

// transportWithAuth adds the Authorization header to each request.
type transportWithAuth struct {
	authToken string
	rt        http.RoundTripper
}

// RoundTrip adds the Authorization header before forwarding the request.
func (t *transportWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "token "+t.authToken)
	return t.rt.RoundTrip(req)
}
