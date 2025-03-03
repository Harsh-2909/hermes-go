// Package models provides implementations of the Model interface, including OpenAI integration.
package models

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/sashabaranov/go-openai"
)

// OpenAIChat implements the Model interface for OpenAI's Chat API.
type OpenAIChat struct {
	client      *openai.Client // Internal OpenAI API client
	Id          string         // Model identifier (e.g., "gpt-4o-mini")
	ApiKey      string         // OpenAI API key for authentication
	Temperature float32        // Controls response randomness; higher values increase creativity
}

// Init initializes the OpenAIChat instance with defaults and validates required fields.
// It panics if ApiKey or Id is missing and sets Temperature to 0.5 if unspecified.
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

func convertMessageToOpenAIFormat(messages []models.Message) []openai.ChatCompletionMessage {
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return openaiMessages
}

// ChatCompletion sends a synchronous chat request to OpenAI and returns the response.
// It converts input messages to OpenAI's format, makes the API call, and constructs a ModelResponse with usage data.
func (model *OpenAIChat) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	openaiMessages := convertMessageToOpenAIFormat(messages)
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

// ChatCompletionStream initiates a streaming chat request to OpenAI and returns a channel of responses.
// It emits ModelResponse events ("chunk" for content, "end" for completion, "error" for failures).
// The caller must consume the channel to process the stream.
func (model *OpenAIChat) ChatCompletionStream(ctx context.Context, messages []models.Message) (chan models.ModelResponse, error) {
	openaiMessages := convertMessageToOpenAIFormat(messages)
	stream, err := model.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       model.Id,
		Messages:    openaiMessages,
		Temperature: model.Temperature,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	ch := make(chan models.ModelResponse)
	go func() {
		defer close(ch)
		for {
			resp, err := stream.Recv()
			// Handle stream errors and completion
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
