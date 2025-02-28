package models

import (
	"context"
	"fmt"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/sashabaranov/go-openai"
)

// OpenAIModel integrates with OpenAI’s Chat API, implementing the Model interface.
type OpenAIModel struct {
	client    *openai.Client // OpenAI API client
	modelName string         // e.g., "gpt-3.5-turbo"
}

// NewOpenAIModel initializes an OpenAIModel with an API key and model name.
func NewOpenAIModel(apiKey, modelName string) *OpenAIModel {
	client := openai.NewClient(apiKey)
	return &OpenAIModel{client: client, modelName: modelName}
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
			Model:    m.modelName,
			Messages: openaiMessages,
		},
	)
	if err != nil {
		return "", err
	}

	// Extract the response content
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from model")
}
