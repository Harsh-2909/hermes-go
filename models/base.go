package models

import (
	"context"
	"time"
)

// Model defines the interface for AI model interactions.
type Model interface {
	Init()
	ChatCompletion(ctx context.Context, messages []Message) (ModelResponse, error)
	ChatCompletionStream(ctx context.Context, messages []Message) (chan ModelResponse, error)
}

// Usage captures token usage or other metrics from the model.
type Usage struct {
	PromptTokens     int // Tokens used in the prompt
	CompletionTokens int // Tokens used in the completion
	TotalTokens      int // Total tokens used
}

// ModelResponse represents a response from the model, used for both streaming and non-streaming cases.
type ModelResponse struct {
	Event     string    // Type of event: "chunk", "tool_call", "end", "complete"
	Data      string    // Response content or chunk for streaming
	Usage     *Usage    // Usage metrics, nullable for partial responses
	CreatedAt time.Time // Timestamp of response creation
	Audio     []byte    // Optional audio data, if applicable
	Thinking  string    // Optional intermediate reasoning or thoughts
}
