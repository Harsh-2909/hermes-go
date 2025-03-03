package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/sashabaranov/go-openai"
)

// TestOpenAIChatInit tests the initialization of the OpenAIChat struct.
func TestOpenAIChatInit(t *testing.T) {
	tests := []struct {
		name        string
		model       OpenAIChat
		shouldPanic bool
		wantTemp    float32
	}{
		{
			name:        "Missing API key",
			model:       OpenAIChat{Id: "test-model"},
			shouldPanic: true,
		},
		{
			name:        "Missing model ID",
			model:       OpenAIChat{ApiKey: "test-key"},
			shouldPanic: true,
		},
		{
			name:        "Valid with default temperature",
			model:       OpenAIChat{ApiKey: "test-key", Id: "test-model"},
			shouldPanic: false,
			wantTemp:    0.5,
		},
		{
			name:        "Valid with custom temperature",
			model:       OpenAIChat{ApiKey: "test-key", Id: "test-model", Temperature: 0.8},
			shouldPanic: false,
			wantTemp:    0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but did not panic")
					}
				}()
			}
			tt.model.Init()
			if !tt.shouldPanic {
				if tt.model.Temperature != tt.wantTemp {
					t.Errorf("Expected Temperature %f, got %f", tt.wantTemp, tt.model.Temperature)
				}
				if tt.model.client == nil {
					t.Errorf("Expected client to be initialized")
				}
			}
		})
	}
}

// TestChatCompletion tests the synchronous ChatCompletion method with a mocked HTTP response.
func TestChatCompletion(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request details
		if r.Method != "POST" || r.URL.Path != "/chat/completions" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("Expected Authorization: Bearer test-key, got %s", auth)
		}

		// Mock response
		resp := openai.ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello, world!",
					},
					FinishReason: "stop",
				},
			},
			Usage: openai.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create OpenAI client with mock server URL
	config := openai.DefaultConfig("test-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	// Initialize OpenAIChat
	model := OpenAIChat{
		client:      client,
		Id:          "gpt-3.5-turbo",
		ApiKey:      "test-key", // Not used since we override client
		Temperature: 0.7,
	}

	// Test ChatCompletion
	ctx := context.Background()
	messages := []models.Message{
		{Role: "user", Content: "Hi there"},
	}
	resp, err := model.ChatCompletion(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Validate response
	if resp.Event != "complete" {
		t.Errorf("Expected Event 'complete', got '%s'", resp.Event)
	}
	if resp.Data != "Hello, world!" {
		t.Errorf("Expected Data 'Hello, world!', got '%s'", resp.Data)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected Usage with TotalTokens 15, got %+v", resp.Usage)
	}
	if resp.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

// TestChatCompletionStream tests the streaming ChatCompletionStream method with a mocked SSE response.
func TestChatCompletionStream(t *testing.T) {
	// Mock server setup for Server-Sent Events (SSE)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter does not support flushing")
		}

		// Simulate streaming response
		fmt.Fprintf(w, "data: %s\n\n", marshalSSE(t, openai.ChatCompletionStreamResponse{
			ID:      "stream-test",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "Hello, ",
					},
				},
			},
		}))
		flusher.Flush()
		time.Sleep(10 * time.Millisecond) // Simulate delay
		fmt.Fprintf(w, "data: %s\n\n", marshalSSE(t, openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "world!",
					},
				},
			},
		}))
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	// Create OpenAI client with mock server URL
	config := openai.DefaultConfig("test-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)

	// Initialize OpenAIChat
	model := OpenAIChat{
		client:      client,
		Id:          "gpt-3.5-turbo",
		ApiKey:      "test-key", // Not used since we override client
		Temperature: 0.7,
	}

	// Test ChatCompletionStream
	ctx := context.Background()
	messages := []models.Message{
		{Role: "user", Content: "Stream me"},
	}
	ch, err := model.ChatCompletionStream(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Consume and validate the stream
	expectedEvents := []string{"chunk", "chunk", "end"}
	expectedData := []string{"Hello, ", "world!", ""}
	i := 0
	for resp := range ch {
		if i >= len(expectedEvents) {
			t.Errorf("Received more events than expected")
			break
		}
		if resp.Event != expectedEvents[i] {
			t.Errorf("Expected event '%s', got '%s'", expectedEvents[i], resp.Event)
		}
		if resp.Data != expectedData[i] {
			t.Errorf("Expected data '%s', got '%s'", expectedData[i], resp.Data)
		}
		if resp.CreatedAt.IsZero() {
			t.Errorf("Expected non-zero CreatedAt for event %d", i)
		}
		i++
	}
	if i != len(expectedEvents) {
		t.Errorf("Expected %d events, got %d", len(expectedEvents), i)
	}
}

// marshalSSE helper function to encode SSE data for the mock server.
func marshalSSE(t *testing.T, resp openai.ChatCompletionStreamResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal SSE response: %v", err)
	}
	return string(data)
}
