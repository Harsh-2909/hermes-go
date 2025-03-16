// models/audio.go
package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Audio represents an audio file provided via URL, file path, or base64 content.
type Audio struct {
	URL      string // URL of the audio file
	FilePath string // Local file path to the audio file
	Base64   string // Base64-encoded audio content
}

// GetType returns the type of the media.
func (a *Audio) GetType() string {
	return "audio"
}

// Content returns the audio data as base64, handling different input types.
func (a *Audio) Content() (string, error) {
	if a.Base64 != "" {
		return a.Base64, nil
	}
	if a.FilePath != "" {
		data, err := os.ReadFile(a.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read audio file: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
	if a.URL != "" {
		resp, err := http.Get(a.URL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch audio from URL: %w", err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read audio data from URL: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return "", fmt.Errorf("no audio data provided")
}
