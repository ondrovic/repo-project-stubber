package httpclient

import (
	"github-project-template/internal/consts"
	"net/http"
)

// Client is a global variable holding the HTTP client used for making requests with authentication support.
var Client *http.Client

type transportWithAuth struct {
	// authToken is the authentication token used for authorized requests.
	authToken string
	// rt is the underlying RoundTripper used for HTTP transport.
	rt http.RoundTripper
}

// InitClient initializes an HTTP client with authentication support using the provided authToken.
// It creates an HTTP client with a custom transport that includes the authentication token for requests.
// Parameters:
// - authToken: The authentication token to be used for authorized requests.
// Returns:
// - A pointer to the initialized http.Client.
// - An error if any issues occur during client initialization (returns nil in this implementation).
func InitClient(authToken string) (*http.Client, error) {
	Client = &http.Client{
		Transport: &transportWithAuth{
			authToken: authToken,
			rt:        http.DefaultTransport,
		},
	}

	return Client, nil
}

// RoundTrip executes a single HTTP request using the transportWithAuth transport.
// If an authentication token is set, it adds an "Authorization" header to the request.
// Parameters:
// - req: The HTTP request to be sent.
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if any issues occur during the request execution.
func (t *transportWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.authToken != consts.EMPTY_STRING {
		req.Header.Set("Authorization", "token "+t.authToken)
	}
	return t.rt.RoundTrip(req)
}
