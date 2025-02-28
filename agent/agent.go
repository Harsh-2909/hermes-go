package agent

import "context"

// Message represents a single message in the conversation.
type Message struct {
	Role    string // e.g., "system", "user", "assistant"
	Content string // The text content of the message
}

// Model defines the interface for AI model interactions.
type Model interface {
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
}

// Agent is the core struct that manages conversation and interacts with a model.
type Agent struct {
	model    Model     // The AI model to use (e.g., OpenAI)
	messages []Message // Conversation history
}

// NewAgent creates a new Agent instance with an optional system message.
// The system message sets the agent’s initial behavior.
func NewAgent(model Model, systemMessage string) *Agent {
	agent := &Agent{
		model:    model,
		messages: []Message{},
	}
	if systemMessage != "" {
		agent.AddMessage("system", systemMessage)
	}
	return agent
}

// AddMessage appends a new message to the conversation history.
func (a *Agent) AddMessage(role, content string) {
	a.messages = append(a.messages, Message{Role: role, Content: content})
}

// RespondTo processes a user message and returns the agent’s response.
// It uses the model to generate a response and updates the conversation history.
func (a *Agent) RespondTo(ctx context.Context, userMessage string) (string, error) {
	a.AddMessage("user", userMessage)
	response, err := a.model.ChatCompletion(ctx, a.messages)
	if err != nil {
		return "", err
	}
	a.AddMessage("assistant", response)
	return response, nil
}
