package models

import (
	"context"
	"fmt"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/sashabaranov/go-openai"
)

// OpenAIModel integrates with OpenAI’s Chat API, implementing the Model interface.
type OpenAIModel struct {
	client      *openai.Client // OpenAI API client
	Id          string         // The model to use, e.g., "gpt-4o-mini"
	ApiKey      string         // The API key for OpenAI
	Temperature float32        // The temperature for the model
}

func (m *OpenAIModel) Init() {
	if m.ApiKey == "" {
		panic("OpenAIModel must have an API key")
	}
	if m.Id == "" {
		panic("OpenAIModel must have a model ID")
	}
	if m.Temperature == 0 {
		m.Temperature = 0.5 // Default temperature
	}
	m.client = openai.NewClient(m.ApiKey)
}

// ChatCompletion sends messages to OpenAI’s Chat API and returns the response.
func (m *OpenAIModel) ChatCompletion(ctx context.Context, messages []agent.Message) (string, error) {
	// Convert our Message type to OpenAI’s expected format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Make the API call
	resp, err := m.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       m.Id,
			Messages:    openaiMessages,
			Temperature: m.Temperature,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get chat completion for model %s: %w", m.Id, err)
	}

	// Extract the response content
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from model")
}
