package models

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
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
func (model *OpenAIChat) Init() {
	if model.ApiKey == "" {
		panic("OpenAIChat must have an API key")
	}
	if model.Id == "" {
		panic("OpenAIChat must have a model ID")
	}
	if model.Temperature == 0 {
		model.Temperature = 0.5 // Default temperature
	}
	model.client = openai.NewClient(model.ApiKey)
}

// ChatCompletion sends messages to OpenAI’s Chat API and returns the response.
func (model *OpenAIChat) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	// Convert our Message type to OpenAI’s expected format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Make the API call
	resp, err := model.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model.Id,
			Messages:    openaiMessages,
			Temperature: model.Temperature,
		},
	)
	if err != nil {
		return models.ModelResponse{}, fmt.Errorf("failed to get chat completion for model %s: %w", model.Id, err)
	}

	// Extract the response content
	if len(resp.Choices) == 0 {
		return models.ModelResponse{}, fmt.Errorf("no response from model")
	}
	modelResp := models.ModelResponse{
		Event:     "complete",
		Data:      resp.Choices[0].Message.Content,
		Usage:     nil,
		CreatedAt: time.Now(),
	}
	modelResp.Usage = &models.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return modelResp, nil
}

func (model *OpenAIChat) ChatCompletionStream(ctx context.Context, messages []models.Message) (chan models.ModelResponse, error) {
	// Convert agent messages to OpenAI format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Create the stream
	stream, err := model.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       model.Id,
		Messages:    openaiMessages,
		Temperature: model.Temperature,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	// Create channel for ModelResponse
	ch := make(chan models.ModelResponse)

	// Process stream in a goroutine
	go func() {
		defer close(ch) // Close channel when done
		for {
			resp, err := stream.Recv()
			// When the stream ends, we get an EOF error
			if err == io.EOF {
				ch <- models.ModelResponse{
					Event:     "end",
					CreatedAt: time.Now(),
				}
				return
			}
			if err != nil {
				ch <- models.ModelResponse{
					Event:     "error",
					Data:      err.Error(),
					CreatedAt: time.Now(),
				}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta
				if delta.Content != "" {
					// Emit content chunk
					ch <- models.ModelResponse{
						Event:     "chunk",
						Data:      delta.Content,
						CreatedAt: time.Now(),
					}
				}
			}
		}
	}()

	return ch, nil
}
