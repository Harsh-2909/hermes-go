package agent

import (
	"context"
	"fmt"
)

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
	Model    Model     // The AI model to use (e.g., OpenAI)
	Messages []Message // Conversation history

	// System message Settings

	SystemMessage string // The initial system message
	Description   string // A description of the Agent that is added to the start of the system message.
	Goal          string // The goal of this task.
	Role          string // The role of the agent in the conversation.
}

func (a *Agent) Init() {
	if a.Model == nil {
		panic("Agent must have a model")
	}
	if a.Messages == nil {
		a.Messages = []Message{}
	}
	a.Messages = append(a.Messages, a.getSystemMessage())
}

func (a *Agent) getSystemMessage() Message {
	var systemMessageContent string
	if a.SystemMessage != "" {
		systemMessageContent = a.SystemMessage
	} else {
		if a.Description != "" {
			systemMessageContent += a.Description + "\n\n"
		}
		if a.Goal != "" {
			systemMessageContent += fmt.Sprintf("<your_goal>\n%s\n</your_goal>\n\n", a.Goal)
		}
		if a.Role != "" {
			systemMessageContent += fmt.Sprintf("<your_role>\n%s\n</your_role>\n\n", a.Role)
		}
	}
	return Message{Role: "system", Content: systemMessageContent}
}

// AddMessage appends a new message to the conversation history.
func (a *Agent) AddMessage(role, content string) {
	a.Messages = append(a.Messages, Message{Role: role, Content: content})
}

// RespondTo processes a user message and returns the agent’s response.
// It uses the model to generate a response and updates the conversation history.
func (a *Agent) RespondTo(ctx context.Context, userMessage string) (string, error) {
	a.AddMessage("user", userMessage)

	if len(a.Messages) == 0 {
		return "", fmt.Errorf("no messages available for chat completion")
	}

	response, err := a.Model.ChatCompletion(ctx, a.Messages)
	if err != nil {
		return "", err
	}
	a.AddMessage("assistant", response)
	return response, nil
}
