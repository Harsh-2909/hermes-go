package agent

import (
	"context"
	"fmt"

	"github.com/Harsh-2909/hermes-go/models"
)

// Agent is the core struct that manages conversation and interacts with a model.
type Agent struct {
	Model    models.Model     // The AI model to use (e.g., OpenAI)
	Messages []models.Message // Conversation history

	// System message Settings

	SystemMessage string // The initial system message
	Description   string // A description of the Agent that is added to the start of the system message.
	Goal          string // The goal of this task.
	Role          string // The role of the agent in the conversation.
}

func (agent *Agent) Init() {
	if agent.Model == nil {
		panic("Agent must have a model")
	}
	if agent.Messages == nil {
		agent.Messages = []models.Message{}
	}
	agent.Messages = append(agent.Messages, agent.getSystemMessage())
}

func (agent *Agent) getSystemMessage() models.Message {
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
	return models.Message{Role: "system", Content: systemMessageContent}
}

// AddMessage appends a new message to the conversation history.
func (agent *Agent) AddMessage(role, content string) {
	agent.Messages = append(agent.Messages, models.Message{Role: role, Content: content})
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
