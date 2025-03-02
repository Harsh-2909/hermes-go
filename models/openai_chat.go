package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/sashabaranov/go-openai"
)

// OpenAIChat integrates with OpenAI’s Chat API, implementing the Model interface.
type OpenAIChat struct {
	client      *openai.Client // OpenAI API client
	Id          string         // The model to use, e.g., "gpt-4o-mini"
	ApiKey      string         // The API key for OpenAI
	Temperature float32        // The temperature for the model
}

// Init initializes the OpenAIChat model. It sets all the defaults and validates the configuration.
func (m *OpenAIChat) Init() {
	if m.ApiKey == "" {
		panic("OpenAIChat must have an API key")
	}
	if m.Id == "" {
		panic("OpenAIChat must have a model ID")
	}
	if m.Temperature == 0 {
		m.Temperature = 0.5 // Default temperature
	}
	m.client = openai.NewClient(m.ApiKey)
}

// ChatCompletion sends messages to OpenAI’s Chat API and returns the response.
func (m *OpenAIChat) ChatCompletion(ctx context.Context, messages []agent.Message) (agent.ModelResponse, error) {
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
		return agent.ModelResponse{}, fmt.Errorf("failed to get chat completion for model %s: %w", m.Id, err)
	}

	// Extract the response content
	if len(resp.Choices) == 0 {
		return agent.ModelResponse{}, fmt.Errorf("no response from model")
	}
	modelResp := agent.ModelResponse{
		Event:     "complete",
		Data:      resp.Choices[0].Message.Content,
		Usage:     nil,
		CreatedAt: time.Now(),
	}
	modelResp.Usage = &agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return modelResp, nil
}
