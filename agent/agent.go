package agent

import (
	"context"
	"fmt"
	"time"
)

// Message represents a single message in the conversation.
type Message struct {
	Role    string // e.g., "system", "user", "assistant"
	Content string // The text content of the message
}

// Model defines the interface for AI model interactions.
type Model interface {
	Init()
	ChatCompletion(ctx context.Context, messages []Message) (ModelResponse, error)
	// ChatCompletionStream(ctx context.Context, messages []Message) (chan ModelResponse, error)
}

// Agent is the core struct that manages conversation and interacts with a model.
type Agent struct {
	Model    Model     // The AI model to use (e.g., OpenAI)
	Messages []Message // Conversation history

	// System message Settings

	SystemMessage string // The initial system message
	Description   string // A description of the Agent that is added to the start of the system message.
	Goal          string // The goal of this task.
	Role          string // The role of the agent in the conversation.
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

func (agent *Agent) Init() {
	if agent.Model == nil {
		panic("Agent must have a model")
	}
	if agent.Messages == nil {
		agent.Messages = []Message{}
	}
	agent.Messages = append(agent.Messages, agent.getSystemMessage())
}

func (agent *Agent) getSystemMessage() Message {
	var systemMessageContent string
	if agent.SystemMessage != "" {
		systemMessageContent = agent.SystemMessage
	} else {
		if agent.Description != "" {
			systemMessageContent += agent.Description + "\n\n"
		}
		if agent.Goal != "" {
			systemMessageContent += fmt.Sprintf("<your_goal>\n%s\n</your_goal>\n\n", agent.Goal)
		}
		if agent.Role != "" {
			systemMessageContent += fmt.Sprintf("<your_role>\n%s\n</your_role>\n\n", agent.Role)
		}
	}
	return Message{Role: "system", Content: systemMessageContent}
}

// AddMessage appends a new message to the conversation history.
func (agent *Agent) AddMessage(role, content string) {
	agent.Messages = append(agent.Messages, Message{Role: role, Content: content})
}

// RespondTo processes a user message and returns the agent’s response.
// It uses the model to generate a response and updates the conversation history.
func (agent *Agent) RespondTo(ctx context.Context, userMessage string) (string, error) {
	agent.AddMessage("user", userMessage)

	if len(agent.Messages) == 0 {
		return "", fmt.Errorf("no messages available for chat completion")
	}

	response, err := agent.Model.ChatCompletion(ctx, agent.Messages)
	if err != nil {
		return "", err
	}
	agent.AddMessage("assistant", response.Data)
	return response.Data, nil
}
