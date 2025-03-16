// models/audio_test.go
package models

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAudio_GetType(t *testing.T) {
	audio := &Audio{}
	if got := audio.GetType(); got != "audio" {
		t.Errorf("Audio.GetType() = %v, want %v", got, "audio")
	}
}

func TestAudio_Content_Base64(t *testing.T) {
	// Test when Base64 is directly provided
	testBase64 := "SGVsbG8gV29ybGQ=" // "Hello World" in base64
	audio := &Audio{
		Base64: testBase64,
	}

	got, err := audio.Content()
	if err != nil {
		t.Errorf("Audio.Content() error = %v, want nil", err)
		return
	}
	if got != testBase64 {
		t.Errorf("Audio.Content() = %v, want %v", got, testBase64)
	}
}

func TestAudio_Content_FilePath(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_audio.mp3")

	testContent := []byte("This is audio file content simulation")
	expectedBase64 := base64.StdEncoding.EncodeToString(testContent)

	// Write test content to the file
	if err := os.WriteFile(tempFile, testContent, 0666); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	audio := &Audio{
		FilePath: tempFile,
	}

	got, err := audio.Content()
	if err != nil {
		t.Errorf("Audio.Content() error = %v, want nil", err)
		return
	}
	if got != expectedBase64 {
		t.Errorf("Audio.Content() = %v, want %v", got, expectedBase64)
	}
}

func TestAudio_Content_URL(t *testing.T) {
	// Create a mock HTTP server
	testContent := []byte("Audio content from URL")
	expectedBase64 := base64.StdEncoding.EncodeToString(testContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	audio := &Audio{
		URL: server.URL,
	}

	got, err := audio.Content()
	if err != nil {
		t.Errorf("Audio.Content() error = %v, want nil", err)
		return
	}
	if got != expectedBase64 {
		t.Errorf("Audio.Content() = %v, want %v", got, expectedBase64)
	}
}

func TestAudio_Content_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		audio   *Audio
		wantErr bool
	}{
		{
			name: "invalid file path",
			audio: &Audio{
				FilePath: "/non/existent/path/audio.mp3",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			audio: &Audio{
				URL: "http://invalid-url-that-does-not-exist.example",
			},
			wantErr: true,
		},
		{
			name:    "no audio data provided",
			audio:   &Audio{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.audio.Content()
			if (err != nil) != tt.wantErr {
				t.Errorf("Audio.Content() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAudio_Content_BadResponse(t *testing.T) {
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

	audio := &Audio{
		URL: server.URL,
	}

	_, err := audio.Content()
	if err == nil {
		t.Error("Audio.Content() expected error for bad response, got nil")
	}
}
