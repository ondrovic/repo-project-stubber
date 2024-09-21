package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	"github-project-template/internal/consts"
	"github-project-template/internal/httpclient"
	"github-project-template/internal/types"
)

// Mock structures for simulating read error
type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

// RoundTripFunc is a function type that implements the http.RoundTripper interface
type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// TestSetColor tests the SetColor function
func TestSetColor(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		input    interface{}
		expected string
	}{
		{"Red string", color.FgRed, "Hello", "\x1b[31mHello\x1b[0m"},
		{"Blue integer", color.FgBlue, 42, "\x1b[34m42\x1b[0m"},
		{"Green float", color.FgGreen, 3.14, "\x1b[32m3.14\x1b[0m"},
		{"Yellow boolean", color.FgYellow, true, "\x1b[33mtrue\x1b[0m"},
		{"Magenta nil", color.FgMagenta, nil, "\x1b[35m<nil>\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetColor(tt.color, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGrabDownloadUrl tests the GrabDownloadUrl function
func TestGrabDownloadUrl(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*httptest.Server)
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedUrl    string
		expectedError  bool
		allowNilClient bool
	}{
		{
			"Successful request",
			func(server *httptest.Server) { httpclient.Client = server.Client() },
			func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(types.GitHubResponse{DownloadURL: "https://example.com/download"})
			},
			"https://example.com/download",
			false,
			false,
		},
		{
			"HTTP request error",
			func(server *httptest.Server) { httpclient.Client = server.Client() },
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
			consts.EMPTY_STRING,
			true,
			false,
		},
		{
			"Invalid JSON response",
			func(server *httptest.Server) { httpclient.Client = server.Client() },
			func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Invalid JSON")) },
			consts.EMPTY_STRING,
			true,
			false,
		},
		{
			"Error reading response body",
			func(server *httptest.Server) {
				httpclient.Client = &http.Client{Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					return &http.Response{StatusCode: 200, Body: &errorReadCloser{}}
				})}
			},
			func(w http.ResponseWriter, r *http.Request) {},
			consts.EMPTY_STRING,
			true,
			false,
		},
		{
			"Empty response body",
			func(server *httptest.Server) { httpclient.Client = server.Client() },
			func(w http.ResponseWriter, r *http.Request) {},
			consts.EMPTY_STRING,
			true,
			false,
		},
		{
			"HTTP client not initialized",
			func(server *httptest.Server) { httpclient.Client = nil },
			func(w http.ResponseWriter, r *http.Request) {},
			consts.EMPTY_STRING,
			true,
			true,
		},
		{
			"Error creating new HTTP request",
			func(server *httptest.Server) { httpclient.Client = server.Client() },
			func(w http.ResponseWriter, r *http.Request) {},
			consts.EMPTY_STRING,
			true,
			false,
		},
		{
			"Error executing HTTP request",
			func(server *httptest.Server) {
				httpclient.Client = &http.Client{Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return nil, errors.New("network error")
					},
				}}
			},
			func(w http.ResponseWriter, r *http.Request) {},
			consts.EMPTY_STRING,
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			originalClient := httpclient.Client
			defer func() { httpclient.Client = originalClient }()

			if tt.setup != nil {
				tt.setup(server)
			}

			if !tt.allowNilClient && httpclient.Client == nil {
				t.Fatalf("HTTP client is nil before test execution")
			}

			var url string
			var err error

			if tt.name == "Error creating new HTTP request" {
				url, err = GrabDownloadUrl("://invalid-url")
			} else {
				url, err = GrabDownloadUrl(server.URL)
			}

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUrl, url)
		})
	}
}

// TestGetReleaseFile tests the GetReleaseFile function
func TestGetReleaseFile(t *testing.T) {
	tests := []struct {
		name            string
		projectLanguage string
		expectedFile    string
		expectedError   bool
	}{
		{"Empty project language", consts.EMPTY_STRING, consts.EMPTY_STRING, false},
		{"Go language", consts.GO_LANG, consts.GORELEASER, false},
		{"Go language (uppercase)", "GO", consts.GORELEASER, false},
		{"Unsupported language", "python", consts.EMPTY_STRING, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := GetReleaseFile(tt.projectLanguage)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "hasn't been implemented yet")
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFile, file)
		})
	}
}

// TestGetVersionFile tests the GetVersionFile function
func TestGetVersionFile(t *testing.T) {
	tests := []struct {
		name            string
		projectLanguage string
		expectedFile    string
		expectedError   bool
	}{
		{"Empty project language", consts.EMPTY_STRING, consts.EMPTY_STRING, false},
		{"Go language", consts.GO_LANG, consts.VERSION_GO, false},
		{"Go language (uppercase)", "GO", consts.VERSION_GO, false},
		{"Unsupported language", "python", consts.EMPTY_STRING, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := GetVersionFile(tt.projectLanguage)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFile, file)
		})
	}
}
