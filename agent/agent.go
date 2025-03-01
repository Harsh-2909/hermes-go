package agent

import "context"

// Message represents a single message in the conversation.
type Message struct {
	Role    string // e.g., "system", "user", "assistant"
	Content string // The text content of the message
}

// Model defines the interface for AI model interactions.
type Model interface {
	Init()
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
}

// Agent is the core struct that manages conversation and interacts with a model.
type Agent struct {
	Model         Model     // The AI model to use (e.g., OpenAI)
	Messages      []Message // Conversation history
	SystemMessage string    // The initial system message
}

func (a *Agent) Init() {
	if a.Model == nil {
		panic("Agent must have a model")
	}
	if a.Messages == nil {
		a.Messages = []Message{}
	}
	if a.SystemMessage != "" {
		a.AddMessage("system", a.SystemMessage)
	}
}

// AddMessage appends a new message to the conversation history.
func (a *Agent) AddMessage(role, content string) {
	a.Messages = append(a.Messages, Message{Role: role, Content: content})
}

// RespondTo processes a user message and returns the agent’s response.
// It uses the model to generate a response and updates the conversation history.
func (a *Agent) RespondTo(ctx context.Context, userMessage string) (string, error) {
	a.AddMessage("user", userMessage)
	response, err := a.Model.ChatCompletion(ctx, a.Messages)
	if err != nil {
		return "", err
	}
	a.AddMessage("assistant", response)
	return response, nil
}
