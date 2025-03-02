// Package agent provides the core functionality for managing AI agent conversations.
package agent

import (
	"context"
	"fmt"

	"github.com/Harsh-2909/hermes-go/models"
)

// Agent manages a conversation with an AI model, maintaining history and settings for the system message.
// It interacts with a provided Model to process user inputs and generate responses.
type Agent struct {
	Model    models.Model     // The AI model used for generating responses (e.g., OpenAIChat)
	Messages []models.Message // History of system, user, and assistant messages in the conversation

	// Settings for building the default system message
	SystemMessage string // Custom system message; if set, overrides other settings
	Description   string // Description of the agent, prepended to the default system message
	Goal          string // Task goal, included in the system message within <your_goal> tags
	Role          string // Agent's role, included in the system message within <your_role> tags
	Markdown      bool   // If true, instructs the model to format responses in Markdown
}

// Init initializes the Agent with required settings and the system message.
// It panics if no Model is provided and ensures Messages is initialized before appending the system message.
func (agent *Agent) Init() {
	if agent.Model == nil {
		panic("Agent must have a model")
	}
	if agent.Messages == nil {
		agent.Messages = []models.Message{}
	}
	agent.Messages = append(agent.Messages, agent.getSystemMessage())
}

// getSystemMessage constructs the initial system message based on the agent's settings.
// It uses SystemMessage if provided; otherwise, it builds a message from Description, Goal, Role,
// and adds Markdown instructions if enabled.
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
	var additionalInformation []string
	if agent.Markdown {
		additionalInformation = append(additionalInformation, "Use markdown to format your answers.")
	}
	if len(additionalInformation) > 0 {
		systemMessageContent += "<additional_information>"
		for _, instruction := range additionalInformation {
			systemMessageContent += fmt.Sprintf("\n- %s", instruction)
		}
		systemMessageContent += "\n</additional_information>\n\n"
	}
	return models.Message{Role: "system", Content: systemMessageContent}
}

// AddMessage appends a new message with the specified role and content to the conversation history.
func (agent *Agent) AddMessage(role, content string) {
	agent.Messages = append(agent.Messages, models.Message{Role: role, Content: content})
}

// Run processes a user message synchronously and returns the model's response.
// It adds the user message to the history, invokes ChatCompletion on the Model, appends the assistant’s response,
// and returns the result. Returns an error if the model fails or no messages exist.
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

// RunStream processes a user message and returns a channel for streaming model responses.
// It adds the user message to the history and invokes ChatCompletionStream on the Model.
// The caller must consume the channel to receive response chunks; the history is not updated here
// due to the streaming nature (see implementation note).
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
