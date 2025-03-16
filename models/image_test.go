// models/image_test.go
package models

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestImage_GetType(t *testing.T) {
	img := &Image{}
	if got := img.GetType(); got != "image" {
		t.Errorf("Image.GetType() = %v, want %v", got, "image")
	}
}

func TestImage_Content_Base64(t *testing.T) {
	// Test when Base64 is directly provided
	testBase64 := "SGVsbG8gV29ybGQ=" // "Hello World" in base64
	img := &Image{
		Base64: testBase64,
	}

	got, err := img.Content()
	if err != nil {
		t.Errorf("Image.Content() error = %v, want nil", err)
		return
	}
	if got != testBase64 {
		t.Errorf("Image.Content() = %v, want %v", got, testBase64)
	}
}

func TestImage_Content_FilePath(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_image.txt")

	testContent := []byte("Hello World")
	expectedBase64 := base64.StdEncoding.EncodeToString(testContent)

	// Write test content to the file
	if err := os.WriteFile(tempFile, testContent, 0666); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	img := &Image{
		FilePath: tempFile,
	}

	got, err := img.Content()
	if err != nil {
		t.Errorf("Image.Content() error = %v, want nil", err)
		return
	}
	if got != expectedBase64 {
		t.Errorf("Image.Content() = %v, want %v", got, expectedBase64)
	}
}

func TestImage_Content_URL(t *testing.T) {
	// Create a mock HTTP server
	testContent := []byte("Hello World from URL")
	expectedBase64 := base64.StdEncoding.EncodeToString(testContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	img := &Image{
		URL: server.URL,
	}

	got, err := img.Content()
	if err != nil {
		t.Errorf("Image.Content() error = %v, want nil", err)
		return
	}
	if got != expectedBase64 {
		t.Errorf("Image.Content() = %v, want %v", got, expectedBase64)
	}
}

func TestImage_Content_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		img     *Image
		wantErr bool
	}{
		{
			name: "invalid file path",
			img: &Image{
				FilePath: "/non/existent/path/image.jpg",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			img: &Image{
				URL: "http://invalid-url-that-does-not-exist.example",
			},
			wantErr: true,
		},
		{
			name:    "no image data provided",
			img:     &Image{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.img.Content()
			if (err != nil) != tt.wantErr {
				t.Errorf("Image.Content() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestImage_Content_BadResponse(t *testing.T) {
	// Create a mock HTTP server that closes the connection unexpectedly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatalf("webserver doesn't support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Failed to hijack connection: %v", err)
		}
		conn.Close()
	}))
	defer server.Close()

	img := &Image{
		URL: server.URL,
	}

	_, err := img.Content()
	if err == nil {
		t.Error("Image.Content() expected error for bad response, got nil")
	}
}
