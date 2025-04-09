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
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
)

// createMockServer creates a mock HTTP server with a custom handler.
func createMockServer(t *testing.T, handlerFunc http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handlerFunc)
	t.Cleanup(server.Close)
	return server
}

// TestClaude_Init tests the Init method of the Claude struct.
func TestClaude_Init(t *testing.T) {
	// Test panic when ApiKey is missing
	assert.Panics(t, func() {
		model := &Claude{Id: "claude-3-sonnet-20240229"}
		model.Init()
	}, "should panic when ApiKey is missing")

	// Test panic when Id is missing
	assert.Panics(t, func() {
		model := &Claude{ApiKey: "test-key"}
		model.Init()
	}, "should panic when Id is missing")

	// Test successful initialization with default values
	model := &Claude{
		ApiKey: "test-key",
		Id:     "claude-3-sonnet-20240229",
	}
	model.Init()
	assert.True(t, model.isInit, "isInit should be true after initialization")
	assert.Equal(t, float32(0.0), model.Temperature, "Temperature should default to 0.0")
	assert.Equal(t, float32(0.0), model.TopP, "TopP should default to 0.0")
	assert.Equal(t, 4096, model.MaxTokens, "MaxTokens should default to 4096")
	assert.NotNil(t, model.client, "client should be initialized")

	// Test initialization with custom values
	model = &Claude{
		ApiKey:      "test-key",
		Id:          "claude-3-sonnet-20240229",
		Temperature: 0.5,
		TopP:        0.9,
		MaxTokens:   1000,
	}
	model.Init()
	assert.Equal(t, float32(0.5), model.Temperature, "Temperature should be set to 0.5")
	assert.Equal(t, float32(0.9), model.TopP, "TopP should be set to 0.9")
	assert.Equal(t, 1000, model.MaxTokens, "MaxTokens should be set to 1000")
}

// TestClaude_SetTools tests the SetTools method of the Claude struct.
func TestClaude_SetTools(t *testing.T) {
	model := &Claude{}
	tools := []tools.Tool{
		{Name: "test-tool", Description: "A test tool"},
	}
	model.SetTools(tools)
	assert.Equal(t, tools, model.tools, "tools should be set correctly")
}

// Test_formatMessages tests the formatMessages function.
func Test_formatMessages(t *testing.T) {
	messages := []models.Message{
		{Role: "system", Content: "System message"},
		{Role: "user", Content: "Hello", Images: []*models.Image{{URL: "http://example.com/image.jpg"}}},
		{Role: "assistant", Content: "Hi there", ToolCalls: []tools.ToolCall{{ID: "call_123", Name: "test-tool", Arguments: `{"param": "value"}`}}},
		{Role: "tool", ToolCallID: "call_123", Content: "Tool result"},
	}

	anthropicMessages, systemMessage, err := formatMessages(messages)
	assert.NoError(t, err, "formatMessages should not return an error")
	assert.Equal(t, "System message", systemMessage, "system message should be extracted correctly")
	assert.Len(t, anthropicMessages, 3, "should have 3 messages (user, assistant, tool)")

	// Check user message
	userMsg := anthropicMessages[0]
	assert.Equal(t, anthropic.MessageParamRoleUser, userMsg.Role, "user message role should be 'user'")
	assert.Len(t, userMsg.Content, 2, "user message should have text and image")
	assert.Equal(t, "Hello", *userMsg.Content[0].GetText(), "user message text should match")
	assert.Equal(t, "http://example.com/image.jpg", *userMsg.Content[1].GetSource().GetURL(), "user message image URL should match")

	// Check assistant message
	assistantMsg := anthropicMessages[1]
	assert.Equal(t, anthropic.MessageParamRoleAssistant, assistantMsg.Role, "assistant message role should be 'assistant'")
	assert.Len(t, assistantMsg.Content, 2, "assistant message should have text and tool_use")
	assert.Equal(t, "Hi there", *assistantMsg.Content[0].GetText(), "assistant message text should match")
	toolUse := assistantMsg.Content[1].OfRequestToolUseBlock
	assert.Equal(t, "call_123", toolUse.ID, "tool call ID should match")
	assert.Equal(t, "test-tool", toolUse.Name, "tool call name should match")
	assert.Equal(t, map[string]interface{}{"param": "value"}, toolUse.Input, "tool call arguments should match")

	// Check tool result message
	toolMsg := anthropicMessages[2]
	assert.Equal(t, anthropic.MessageParamRoleUser, toolMsg.Role, "tool message role should be 'user'")
	assert.Len(t, toolMsg.Content, 1, "tool message should have one content block")
	toolResult := toolMsg.Content[0].OfRequestToolResultBlock
	assert.Equal(t, "call_123", toolResult.ToolUseID, "tool result ID should match")
	assert.Equal(t, "Tool result", *toolResult.Content[0].GetText(), "tool result content should match")
}

// TestClaude_ChatCompletion tests the ChatCompletion method of the Claude struct.
func TestClaude_ChatCompletion(t *testing.T) {
	// Mock server for synchronous response
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "expected POST method")
		assert.Equal(t, "/v1/messages", r.URL.Path, "expected path /v1/messages")

		// Mock response
		resp := anthropic.Message{
			ID:   "msg_123",
			Role: "assistant",
			Content: []anthropic.ContentBlockUnion{
				{Type: "text", Text: "Hello, world!"},
			},
			Usage: anthropic.Usage{
				InputTokens:  10,
				OutputTokens: 5,
			},
			StopReason: "end_turn",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Configure client to use mock server
	fmt.Println(server.URL)
	client := anthropic.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	model := &Claude{
		ApiKey: "test-key",
		Id:     "claude-3-sonnet-20240229",
		client: &client,
	}
	model.Init()

	messages := []models.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := model.ChatCompletion(context.Background(), messages)
	assert.NoError(t, err, "ChatCompletion should not return an error")
	assert.Equal(t, "complete", resp.Event, "event should be 'complete'")
	assert.Equal(t, "Hello, world!", resp.Data, "response data should match")
	assert.Nil(t, resp.ToolCalls, "no tool calls should be present")
	assert.Equal(t, 10, resp.Usage.PromptTokens, "prompt tokens should match")
	assert.Equal(t, 5, resp.Usage.CompletionTokens, "completion tokens should match")
	assert.Equal(t, 15, resp.Usage.TotalTokens, "total tokens should match")
}

// TestClaude_ChatCompletionStream tests the ChatCompletionStream method of the Claude struct.
func TestClaude_ChatCompletionStream(t *testing.T) {
	// Mock server for streaming response
	server := createMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "expected POST method")
		assert.Equal(t, "/v1/messages", r.URL.Path, "expected path /v1/messages")

		// Simulate streaming events
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		events := []string{
			`event: message_start
data: {"type": "message_start", "message": {"id": "msg_123", "role": "assistant", "usage": {"input_tokens": 10, "output_tokens": 0}}}`,

			`event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": "Hello, "}}`,

			`event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "world!"}}`,

			`event: content_block_stop
data: {"type": "content_block_stop", "index": 0}`,

			`event: message_delta
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn"}, "usage": {"output_tokens": 5}}`,

			`event: message_stop
data: {"type": "message_stop"}`,
		}

		for _, event := range events {
			fmt.Fprint(w, event+"\n\n")
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond) // Simulate delay
		}
	})

	// Configure client to use mock server
	client := anthropic.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(server.URL),
	)

	model := &Claude{
		ApiKey: "test-key",
		Id:     "claude-3-sonnet-20240229",
		client: &client,
	}
	model.Init()

	messages := []models.Message{
		{Role: "user", Content: "Hello"},
	}

	ch, err := model.ChatCompletionStream(context.Background(), messages)
	assert.NoError(t, err, "ChatCompletionStream should not return an error")

	// Collect responses from the channel
	var responses []models.ModelResponse
	for resp := range ch {
		responses = append(responses, resp)
	}

	// Expected events: chunk ("Hello, "), chunk ("world!"), end
	assert.Len(t, responses, 3, "should receive 3 events: 2 chunks and 1 end")
	assert.Equal(t, "chunk", responses[0].Event, "first event should be 'chunk'")
	assert.Equal(t, "Hello, ", responses[0].Data, "first chunk data should match")
	assert.Equal(t, "chunk", responses[1].Event, "second event should be 'chunk'")
	assert.Equal(t, "world!", responses[1].Data, "second chunk data should match")
	assert.Equal(t, "end", responses[2].Event, "third event should be 'end'")
	assert.Nil(t, responses[2].ToolCalls, "no tool calls should be present")
	assert.Equal(t, 5, responses[2].Usage.CompletionTokens, "completion tokens should match in end event")
}
