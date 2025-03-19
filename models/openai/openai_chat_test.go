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
	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
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
			wantTemp:    0.0,
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

// TestConvertMessageToOpenAIFormat tests the conversion of messages to OpenAI format.
func TestConvertMessageToOpenAIFormat(t *testing.T) {
	messages := []models.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
		{Role: "user", Content: "Describe this image", Images: []*models.Image{{URL: "http://example.com/image.png"}}},
		{Role: "user", Content: "Describe this image", Audios: []*models.Audio{{URL: "http://example.com/audio.wav"}}},
	}
	openaiMessages, err := convertMessageToOpenAIFormat(messages)
	assert.NoError(t, err)
	assert.Len(t, openaiMessages, 4)

	assert.Equal(t, "user", openaiMessages[0].Role)
	assert.Equal(t, "Hello", openaiMessages[0].Content)

	assert.Equal(t, "assistant", openaiMessages[1].Role)
	assert.Equal(t, "Hi there", openaiMessages[1].Content)

	assert.Equal(t, "user", openaiMessages[2].Role)
	assert.Len(t, openaiMessages[2].MultiContent, 2)

	assert.Equal(t, "user", openaiMessages[3].Role)
	assert.Len(t, openaiMessages[3].MultiContent, 2)
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
	assert.NoError(t, err)

	// Validate response
	assert.Equal(t, "complete", resp.Event)
	assert.Equal(t, "Hello, world!", resp.Data)
	assert.NotNil(t, resp.Usage)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
	assert.False(t, resp.CreatedAt.IsZero())
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
	assert.NoError(t, err)

	// Consume and validate the stream
	expectedEvents := []string{"chunk", "chunk", "end"}
	expectedData := []string{"Hello, ", "world!", ""}
	i := 0
	for resp := range ch {
		if i >= len(expectedEvents) {
			t.Errorf("Received more events than expected")
			break
		}
		assert.Equal(t, expectedEvents[i], resp.Event)
		assert.Equal(t, expectedData[i], resp.Data)
		assert.False(t, resp.CreatedAt.IsZero())
		i++
	}
	assert.Len(t, expectedEvents, i)
}

// TestChatCompletionWithToolCalls tests the synchronous ChatCompletion method with tool calls.
func TestChatCompletionWithToolCalls(t *testing.T) {
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

		// Mock response with tool calls
		resp := openai.ChatCompletionResponse{
			ID:      "chatcmpl-tool",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "I will use the calculator tool.",
						ToolCalls: []openai.ToolCall{
							{
								ID:   "tool-1",
								Type: "function",
								Function: openai.FunctionCall{
									Name:      "calculate",
									Arguments: `{"a": 5, "b": 3}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
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

	// Initialize OpenAIChat with tools
	model := OpenAIChat{
		client:      client,
		Id:          "gpt-3.5-turbo",
		ApiKey:      "test-key", // Not used since we override client
		Temperature: 0.7,
	}
	model.SetTools([]tools.Tool{
		{
			Name:        "calculate",
			Description: "Calculates a sum",
			Parameters:  map[string]interface{}{},
			Execute: func(ctx context.Context, args string) (string, error) {
				return "8", nil
			},
		},
	})

	// Test ChatCompletion with tool calls
	ctx := context.Background()
	messages := []models.Message{
		{Role: "user", Content: "What is 5 + 3?"},
	}
	resp, err := model.ChatCompletion(ctx, messages)
	assert.NoError(t, err)

	// Validate response
	assert.Equal(t, "tool_call", resp.Event)
	assert.Equal(t, "I will use the calculator tool.", resp.Data)
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "tool-1", resp.ToolCalls[0].ID)
	assert.Equal(t, "calculate", resp.ToolCalls[0].Name)
	assert.Equal(t, `{"a": 5, "b": 3}`, resp.ToolCalls[0].Arguments)
	assert.NotNil(t, resp.Usage)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
	assert.False(t, resp.CreatedAt.IsZero())
}

// TestChatCompletionStreamWithToolCalls tests the streaming ChatCompletionStream method with tool calls.
func TestChatCompletionStreamWithToolCalls(t *testing.T) {
	// Mock server setup for Server-Sent Events (SSE)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter does not support flushing")
		}

		// Simulate streaming response with tool call deltas
		fmt.Fprintf(w, "data: %s\n\n", marshalSSE(t, openai.ChatCompletionStreamResponse{
			ID:      "stream-tool",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "Let me calculate that for you. ",
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
						ToolCalls: []openai.ToolCall{
							{
								Index:    ptr(0),
								ID:       "tool-1",
								Type:     "function",
								Function: openai.FunctionCall{Name: "calculate"},
							},
						},
					},
				},
			},
		}))
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)
		fmt.Fprintf(w, "data: %s\n\n", marshalSSE(t, openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						ToolCalls: []openai.ToolCall{
							{
								Index:    ptr(0),
								Function: openai.FunctionCall{Arguments: `{"a": 5, `},
							},
						},
					},
				},
			},
		}))
		flusher.Flush()
		time.Sleep(10 * time.Millisecond)
		fmt.Fprintf(w, "data: %s\n\n", marshalSSE(t, openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						ToolCalls: []openai.ToolCall{
							{
								Index:    ptr(0),
								Function: openai.FunctionCall{Arguments: `"b": 3}`},
							},
						},
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

	// Initialize OpenAIChat with tools
	model := OpenAIChat{
		client:      client,
		Id:          "gpt-3.5-turbo",
		ApiKey:      "test-key", // Not used since we override client
		Temperature: 0.7,
	}
	model.SetTools([]tools.Tool{
		{
			Name:        "calculate",
			Description: "Calculates a sum",
			Parameters:  map[string]interface{}{},
			Execute: func(ctx context.Context, args string) (string, error) {
				return "8", nil
			},
		},
	})

	// Test ChatCompletionStream with tool calls
	ctx := context.Background()
	messages := []models.Message{
		{Role: "user", Content: "What is 5 + 3?"},
	}
	ch, err := model.ChatCompletionStream(ctx, messages)
	assert.NoError(t, err)

	// Consume and validate the stream
	expectedEvents := []string{"chunk", "tool_call", "end"}
	expectedData := []string{"Let me calculate that for you. ", "Let me calculate that for you. ", ""}
	expectedToolCalls := [][]tools.ToolCall{
		nil,
		{{ID: "tool-1", Name: "calculate", Arguments: `{"a": 5, "b": 3}`}},
		nil,
	}
	i := 0
	for resp := range ch {
		if i >= len(expectedEvents) {
			t.Errorf("Received more events than expected")
			break
		}
		assert.Equal(t, expectedEvents[i], resp.Event)
		assert.Equal(t, expectedData[i], resp.Data)
		if expectedToolCalls[i] != nil {
			assert.Len(t, resp.ToolCalls, len(expectedToolCalls[i]))
			for j, tc := range resp.ToolCalls {
				assert.Equal(t, expectedToolCalls[i][j].ID, tc.ID)
				assert.Equal(t, expectedToolCalls[i][j].Name, tc.Name)
				assert.Equal(t, expectedToolCalls[i][j].Arguments, tc.Arguments)
			}
		} else {
			assert.Empty(t, resp.ToolCalls)
		}
		assert.False(t, resp.CreatedAt.IsZero())
		i++
	}
	assert.Len(t, expectedEvents, i)
}

// Helper function to create a pointer to an int
func ptr(i int) *int {
	return &i
}

// marshalSSE helper function to encode SSE data for the mock server.
func marshalSSE(t *testing.T, resp openai.ChatCompletionStreamResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal SSE response: %v", err)
	}
	return string(data)
}
