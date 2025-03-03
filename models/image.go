// models/image.go
package models

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Image represents an image provided via URL, file path, or base64 content.
type Image struct {
	URL      string // URL of the image
	FilePath string // Local file path of the image
	Base64   string // Base64-encoded image content
}

// Content returns the image content as base64, handling different input types.
func (img *Image) Content() (string, error) {
	// If base64 is provided directly, return it
	if img.Base64 != "" {
		return img.Base64, nil
	}
	// If a file path is provided, read and encode the file
	if img.FilePath != "" {
		data, err := ioutil.ReadFile(img.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
	// If a URL is provided, fetch and encode the image
	if img.URL != "" {
		resp, err := http.Get(img.URL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read image data: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return "", fmt.Errorf("no image data provided")
}
