// models/image.go
package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Image represents an image provided via URL, file path, or base64 content.
type Image struct {
	URL      string // URL of the image
	FilePath string // Local file path of the image
	Base64   string // Base64-encoded image content
}

// GetType returns the type of the media.
func (img *Image) GetType() string {
	return "image"
}

// Content returns the image content as base64, handling different input types.
func (img *Image) Content() (string, error) {
	// If base64 is provided directly, return it
	if img.Base64 != "" {
		return img.Base64, nil
	}
	// If a file path is provided, read and encode the file
	if img.FilePath != "" {
		data, err := os.ReadFile(img.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		img.Base64 = base64.StdEncoding.EncodeToString(data)
		return img.Base64, nil
	}
	// If a URL is provided, fetch and encode the image
	if img.URL != "" {
		resp, err := http.Get(img.URL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read image data: %w", err)
		}
		img.Base64 = base64.StdEncoding.EncodeToString(data)
		return img.Base64, nil
	}
	return "", fmt.Errorf("no image data provided")
}

// GetMediaType returns the media type (e.g., image/jpeg, image/png) of the image based on base64 content.
func (img *Image) GetMediaType() (string, error) {
	base64Content, err := img.Content()
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 content: %w", err)
	}
	// http.DetectContentType reads up to 512 bytes to determine the content type.
	mediaType := http.DetectContentType(data)
	return mediaType, nil
}
