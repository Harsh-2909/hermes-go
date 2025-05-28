package utils

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// MakeAPICall performs an HTTP request with the provided parameters.
// It returns the HTTP status code, the response body, and an error if any.
func MakeAPICall(ctx context.Context, method, url string, headers map[string]string, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(respBody), nil
}
