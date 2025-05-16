package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeAPICall_GET(t *testing.T) {
	// Test a GET request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, GET"))
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	ctx := context.Background()
	status, resp, err := MakeAPICall(ctx, http.MethodGet, server.URL, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Hello, GET", resp)
}

func TestMakeAPICall_POST(t *testing.T) {
	// Test a POST request with header and body verification
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if r.Header.Get("X-Test-Header") != "test" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			body, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Posted: " + string(body)))
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	ctx := context.Background()
	headers := map[string]string{"X-Test-Header": "test"}
	status, resp, err := MakeAPICall(ctx, http.MethodPost, server.URL, headers, "data")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, "Posted: data", resp)
}

func TestMakeAPICall_InvalidURL(t *testing.T) {
	// Test with an invalid URL to trigger an error
	ctx := context.Background()
	status, resp, err := MakeAPICall(ctx, http.MethodGet, "://invalid-url", nil, "")
	assert.Error(t, err)
	assert.Equal(t, 0, status)
	assert.Equal(t, "", resp)
}
