// Package models defines shared types and interfaces for AI model interactions.
package models

import (
	"github.com/Harsh-2909/hermes-go/tools"
)

// Message represents a single entry in a conversation with an AI model.
// TODO: Create constants for roles
type Message struct {
	Role       string           // Role of the sender: "system" (instructions), "user" (input), "assistant" (response) or "tool" (tool response)
	Content    string           // Text content of the message
	ToolCallID string           // Unique ID for the tool call (used in OpenAI's API)
	ToolCalls  []tools.ToolCall // Tool calls to execute, this field stores the request with results in the conversion history.

	// Additional Modalities

	Images []*Image // Images attached to the message
	Audios []*Audio // Audio files attached to the message
}
