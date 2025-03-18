// Package models defines shared types and interfaces for AI model interactions.
package models

import (
	"context"
	"time"

	"github.com/Harsh-2909/hermes-go/tools"
)

// Model defines the interface for interacting with AI models.
// Implementations must support initialization and both synchronous and streaming chat completions.
type Model interface {
	Init()                                                                                    // Initialize the model with defaults and validate configuration
	SetTools(tools []tools.Tool)                                                              // Set tools for the model
	ChatCompletion(ctx context.Context, messages []Message) (ModelResponse, error)            // Perform a synchronous chat completion
	ChatCompletionStream(ctx context.Context, messages []Message) (chan ModelResponse, error) // Stream chat responses
}

// Usage captures token usage metrics returned by the model.
type Usage struct {
	PromptTokens     int // Number of tokens in the input prompt
	CompletionTokens int // Number of tokens in the generated completion
	TotalTokens      int // Total tokens used (prompt + completion)
}

// ModelResponse represents a response from an AI model.
// It is used for both synchronous responses (Event="complete") and streaming chunks (e.g., Event="chunk", "end").
type ModelResponse struct {
	Event     string           // Event type: "chunk" (partial data), "complete" (full response), "end" (stream end), "tool_call" (tool execution), etc.
	Data      string           // Response content or chunk data
	Usage     *Usage           // Token usage metrics, typically set for "complete" or "end" events; nullable
	CreatedAt time.Time        // Timestamp when the response was generated
	Audio     []byte           // Optional audio data, if supported by the model
	Thinking  string           // Optional intermediate reasoning or thoughts, if provided
	ToolCalls []tools.ToolCall // Optional tool calls to execute, if provided by the model
}

// Media represents a media object (e.g., text, image, audio) that can be processed by AI models.
type Media interface {
	GetType() string
	Content() (string, error)
}
