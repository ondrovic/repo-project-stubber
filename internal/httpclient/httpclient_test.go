package httpclient

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitClient(t *testing.T) {
	// Arrange
	authToken := "test-auth-token"
	expectedTransport := transportWithAuth{
		authToken: authToken,
		rt:        http.DefaultTransport,
	}

	// Act
	client, err := InitClient(authToken)

	// Assert
	require.NoError(t, err, "InitClient should not return an error")
	require.NotNil(t, client, "InitClient should return a non-nil client")
	assert.IsType(t, &http.Client{}, client, "InitClient should return an *http.Client")
	assert.IsType(t, &transportWithAuth{}, client.Transport, "Client transport should be of type *transportWithAuth")

	// Type assertion to check transport properties
	actualTransport, ok := client.Transport.(*transportWithAuth)
	require.True(t, ok, "Transport should be of type *transportWithAuth")
	assert.Equal(t, expectedTransport.authToken, actualTransport.authToken, "Auth token should match")
}

func TestRoundTrip_WithAuthToken(t *testing.T) {
	// Arrange
	authToken := "test-auth-token"
	InitClient(authToken)

	// Create a test server that echoes back the request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token "+authToken, r.Header.Get("Authorization"), "Authorization header should be set correctly")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a request to the test server
	req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
	require.NoError(t, err, "Failed to create request")

	// Act
	resp, err := Client.Do(req)

	// Assert
	require.NoError(t, err, "Client.Do should not return an error")
	require.NotNil(t, resp, "Client.Do should return a non-nil response")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status code should be OK")
}

func TestRoundTrip_WithoutAuthToken(t *testing.T) {
	// Arrange
	InitClient("") // Initialize with an empty auth token

	// Create a test server that echoes back the request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"), "Authorization header should not be set")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a request to the test server
	req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
	require.NoError(t, err, "Failed to create request")

	// Act
	resp, err := Client.Do(req)

	// Assert
	require.NoError(t, err, "Client.Do should not return an error")
	require.NotNil(t, resp, "Client.Do should return a non-nil response")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status code should be OK")
}
