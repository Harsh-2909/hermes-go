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

// Run processes a user message and returns the agent’s response.
// It uses the model to generate a response and updates the conversation history.
func (agent *Agent) Run(ctx context.Context, userMessage string) (models.ModelResponse, error) {
	agent.AddMessage("user", userMessage)

	if len(agent.Messages) == 0 {
		return models.ModelResponse{}, fmt.Errorf("no messages available for chat completion")
	}

	response, err := agent.Model.ChatCompletion(ctx, agent.Messages)
	if err != nil {
		return models.ModelResponse{}, err
	}
	agent.AddMessage("assistant", response.Data)
	return response, nil
}

// RunStream processes a user message and returns a channel of responses.
// It uses the model to generate a response and updates the conversation history.
func (agent *Agent) RunStream(ctx context.Context, userMessage string) (chan models.ModelResponse, error) {
	agent.AddMessage("user", userMessage)

	if len(agent.Messages) == 0 {
		return nil, fmt.Errorf("no messages available for chat completion")
	}

	ch, err := agent.Model.ChatCompletionStream(ctx, agent.Messages)
	if err != nil {
		return nil, err
	}

	// Accumulate response in the background for history.
	// FIXME: This implementation does not work properly as it is consuming the channel before the caller of `RunStream` can consume it.
	// fullResponse := ""
	// go func() {
	// 	for resp := range ch {
	// 		if resp.Event == "chunk" {
	// 			fullResponse += resp.Data
	// 		}
	// 		// Add handling for other events (e.g., tool calls) as needed
	// 	}
	// 	agent.AddMessage("assistant", fullResponse)
	// 	fmt.Println("DEBUGGING. Full response:", fullResponse)
	// }()

	return ch, nil
}
