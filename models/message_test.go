package models

import (
	"testing"
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

			if msg.Role != tt.wantRole {
				t.Errorf("Message.Role = %v, want %v", msg.Role, tt.wantRole)
			}
			if msg.Content != tt.content {
				t.Errorf("Message.Content = %v, want %v", msg.Content, tt.content)
			}
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

	if msg.Content != content {
		t.Errorf("Message.Content = %v, want %v", msg.Content, content)
	}
	if len(msg.Images) != 0 {
		t.Errorf("Message.Images should be nil or empty, got %v", msg.Images)
	}
	if len(msg.Audios) != 0 {
		t.Errorf("Message.Audios should be nil or empty, got %v", msg.Audios)
	}
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

	if len(msg.Images) != 3 {
		t.Errorf("Message.Images length = %d, want %d", len(msg.Images), 3)
	}

	// Verify each image is correctly stored
	if msg.Images[0].URL != "https://example.com/image1.jpg" {
		t.Errorf("Message.Images[0].URL = %v, want %v", msg.Images[0].URL, "https://example.com/image1.jpg")
	}
	if msg.Images[1].FilePath != "/path/to/image2.png" {
		t.Errorf("Message.Images[1].FilePath = %v, want %v", msg.Images[1].FilePath, "/path/to/image2.png")
	}
	if msg.Images[2].Base64 != "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=" {
		t.Errorf("Message.Images[2].Base64 incorrect")
	}
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

	if len(msg.Audios) != 3 {
		t.Errorf("Message.Audios length = %d, want %d", len(msg.Audios), 3)
	}

	// Verify each audio is correctly stored
	if msg.Audios[0].URL != "https://example.com/audio1.mp3" {
		t.Errorf("Message.Audios[0].URL = %v, want %v", msg.Audios[0].URL, "https://example.com/audio1.mp3")
	}
	if msg.Audios[1].FilePath != "/path/to/audio2.wav" {
		t.Errorf("Message.Audios[1].FilePath = %v, want %v", msg.Audios[1].FilePath, "/path/to/audio2.wav")
	}
	if msg.Audios[2].Base64 != "UklGRiQAAABXQVZFZm10IBAAAAABAAEARKwAAIhYAQACABAAZGF0YQAAAAA=" {
		t.Errorf("Message.Audios[2].Base64 incorrect")
	}
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

	if len(msg.Images) != 1 {
		t.Errorf("Message.Images length = %d, want %d", len(msg.Images), 1)
	}

	if len(msg.Audios) != 1 {
		t.Errorf("Message.Audios length = %d, want %d", len(msg.Audios), 1)
	}

	if msg.Images[0].URL != "https://example.com/image.jpg" {
		t.Errorf("Message.Images[0].URL = %v, want %v", msg.Images[0].URL, "https://example.com/image.jpg")
	}

	if msg.Audios[0].URL != "https://example.com/audio.mp3" {
		t.Errorf("Message.Audios[0].URL = %v, want %v", msg.Audios[0].URL, "https://example.com/audio.mp3")
	}
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
				if len(msg.Images) != 0 {
					t.Errorf("Expected empty Images array, got %v", msg.Images)
				}
				if len(msg.Audios) != 0 {
					t.Errorf("Expected empty Audios array, got %v", msg.Audios)
				}
			}
		})
	}
}
