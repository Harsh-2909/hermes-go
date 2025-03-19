package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageWithDifferentRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		content  string
		wantRole string
	}{
		{
			name:     "System role",
			role:     "system",
			content:  "You are a helpful assistant",
			wantRole: "system",
		},
		{
			name:     "User role",
			role:     "user",
			content:  "Hello, can you help me?",
			wantRole: "user",
		},
		{
			name:     "Assistant role",
			role:     "assistant",
			content:  "Yes, I'd be happy to help you!",
			wantRole: "assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:    tt.role,
				Content: tt.content,
			}

			assert.Equal(t, tt.wantRole, msg.Role)
			assert.Equal(t, tt.content, msg.Content)
		})
	}
}

func TestMessageWithTextOnly(t *testing.T) {
	content := "This is a text-only message"
	msg := Message{
		Role:    "user",
		Content: content,
		Images:  nil,
		Audios:  nil,
	}

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, content, msg.Content)
	assert.Nil(t, msg.Images)
	assert.Nil(t, msg.Audios)
}

func TestMessageWithImages(t *testing.T) {
	images := []*Image{
		{
			URL: "https://example.com/image1.jpg",
		},
		{
			FilePath: "/path/to/image2.png",
		},
		{
			Base64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=",
		},
	}

	msg := Message{
		Role:    "user",
		Content: "Check out these images",
		Images:  images,
	}

	// Verify the number of images
	assert.Len(t, msg.Images, 3)

	// Verify each image is correctly stored
	assert.Equal(t, "https://example.com/image1.jpg", msg.Images[0].URL)
	assert.Equal(t, "/path/to/image2.png", msg.Images[1].FilePath)
	assert.Equal(t, "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=", msg.Images[2].Base64)
}

func TestMessageWithAudio(t *testing.T) {
	audios := []*Audio{
		{
			URL: "https://example.com/audio1.mp3",
		},
		{
			FilePath: "/path/to/audio2.wav",
		},
		{
			Base64: "UklGRiQAAABXQVZFZm10IBAAAAABAAEARKwAAIhYAQACABAAZGF0YQAAAAA=", // Minimal WAV file
		},
	}

	msg := Message{
		Role:    "user",
		Content: "Listen to these audio clips",
		Audios:  audios,
	}

	// Verify the number of audio clips
	assert.Len(t, msg.Audios, 3)

	// Verify each audio is correctly stored
	assert.Equal(t, "https://example.com/audio1.mp3", msg.Audios[0].URL)
	assert.Equal(t, "/path/to/audio2.wav", msg.Audios[1].FilePath)
	assert.Equal(t, "UklGRiQAAABXQVZFZm10IBAAAAABAAEARKwAAIhYAQACABAAZGF0YQAAAAA=", msg.Audios[2].Base64)
}

func TestMessageWithImagesAndAudio(t *testing.T) {
	images := []*Image{
		{
			URL: "https://example.com/image.jpg",
		},
	}

	audios := []*Audio{
		{
			URL: "https://example.com/audio.mp3",
		},
	}

	msg := Message{
		Role:    "assistant",
		Content: "Here's the image and audio you requested",
		Images:  images,
		Audios:  audios,
	}

	assert.NotNil(t, msg.Images)
	assert.NotNil(t, msg.Audios)

	assert.Len(t, msg.Images, 1)
	assert.Len(t, msg.Audios, 1)

	assert.Equal(t, "https://example.com/image.jpg", msg.Images[0].URL)
	assert.Equal(t, "https://example.com/audio.mp3", msg.Audios[0].URL)
}

func TestMessageWithEmptyArrays(t *testing.T) {
	tests := []struct {
		name        string
		images      []*Image
		audios      []*Audio
		expectEmpty bool
	}{
		{
			name:        "Nil arrays",
			images:      nil,
			audios:      nil,
			expectEmpty: true,
		},
		{
			name:        "Empty arrays",
			images:      []*Image{},
			audios:      []*Audio{},
			expectEmpty: true,
		},
		{
			name:        "Nil images, empty audios",
			images:      nil,
			audios:      []*Audio{},
			expectEmpty: true,
		},
		{
			name:        "Empty images, nil audios",
			images:      []*Image{},
			audios:      nil,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:    "user",
				Content: "A message with empty arrays",
				Images:  tt.images,
				Audios:  tt.audios,
			}

			if tt.expectEmpty {
				assert.Len(t, msg.Images, 0)
				assert.Len(t, msg.Audios, 0)
			} else {
				assert.NotNil(t, msg.Images)
				assert.NotNil(t, msg.Audios)
			}
		})
	}
}
