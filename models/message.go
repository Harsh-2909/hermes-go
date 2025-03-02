package models

// Message represents a single message in the conversation.
type Message struct {
	Role    string // e.g., "system", "user", "assistant"
	Content string // The text content of the message
}
