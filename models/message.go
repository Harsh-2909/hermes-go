// Package models defines shared types and interfaces for AI model interactions.
package models

// Message represents a single entry in a conversation with an AI model.
type Message struct {
	Role    string // Role of the sender: "system" (instructions), "user" (input), or "assistant" (response)
	Content string // Text content of the message
}
