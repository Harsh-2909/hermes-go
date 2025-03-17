// Package models provides implementations of the Model interface, including OpenAI integration.
package models

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/Harsh-2909/hermes-go/utils"
	"github.com/sashabaranov/go-openai"
)

// OpenAIChat implements the Model interface for OpenAI's Chat API.
type OpenAIChat struct {
	ApiKey           string  // Required OpenAI API key.
	Id               string  // Required model ID (e.g., "gpt-4o-mini")
	Temperature      float32 // In [0,2] range. Higher values -> more creative.
	PresencePenalty  float32 // In [-2,2] range.
	FrequencyPenalty float32 // In [-2,2] range.
	Stop             []string
	N                int
	User             string
	// An alternative to sampling with temperature, called nucleus sampling.
	// The model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP float32
	// MaxCompletionTokens An upper bound for the number of tokens that can be generated for a completion,
	// including visible output tokens and reasoning tokens https://platform.openai.com/docs/guides/reasoning
	MaxCompletionTokens int
	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool
	// TopLogProbs is an integer between 0 and 20 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int

	// Internal fields

	client *openai.Client // Internal OpenAI API client
	isInit bool           // Internal flag to track initialization
	tools  []tools.Tool   // Internal list of tools
}

// Init initializes the OpenAIChat instance with defaults and validates required fields.
// It panics if ApiKey or Id is missing.
func (model *OpenAIChat) Init() {
	if model.isInit {
		return
	}
	if model.ApiKey == "" {
		panic("OpenAIChat must have an API key")
	}
	if model.Id == "" {
		panic("OpenAIChat must have a model ID")
	}
	if model.Temperature < 0 || model.Temperature > 2 {
		model.Temperature = 0.5
	}
	if model.TopP < 0 || model.TopP > 1 {
		model.TopP = 1.0
	}
	if model.MaxCompletionTokens < 0 {
		model.MaxCompletionTokens = 0
	}
	if model.PresencePenalty < -2 || model.PresencePenalty > 2 {
		model.PresencePenalty = 0
	}
	if model.FrequencyPenalty < -2 || model.FrequencyPenalty > 2 {
		model.FrequencyPenalty = 0
	}
	if model.TopLogProbs < 0 || model.TopLogProbs > 20 {
		model.TopLogProbs = 0
	}
	if model.N < 1 {
		model.N = 1
	}

	model.client = openai.NewClient(model.ApiKey)
	model.isInit = true
}

func (model *OpenAIChat) SetTools(tools []tools.Tool) {
	model.tools = tools
}

// convertMessageToOpenAIFormat converts a slice of Message instances to OpenAI's ChatCompletionMessage format.
// It handles text and image content, converting images to base64-encoded URLs.
func convertMessageToOpenAIFormat(messages []models.Message) ([]openai.ChatCompletionMessage, error) {
	var openaiMessages []openai.ChatCompletionMessage
	var chatMessage openai.ChatCompletionMessage
	for _, msg := range messages {
		var contentParts []openai.ChatMessagePart
		if len(msg.Images) > 0 || len(msg.Audios) > 0 {
			if msg.Content != "" {
				contentParts = append(contentParts, openai.ChatMessagePart{
					Type: "text",
					Text: msg.Content,
				})
			}
			for _, img := range msg.Images {
				base64Content, err := img.Content()
				if err != nil {
					return nil, fmt.Errorf("failed to get image content: %w", err)
				}
				contentParts = append(contentParts, openai.ChatMessagePart{
					Type: "image_url",
					ImageURL: &openai.ChatMessageImageURL{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Content),
					},
				})
			}
			for _, audio := range msg.Audios {
				base64Content, err := audio.Content()
				if err != nil {
					return nil, fmt.Errorf("failed to get audio content: %w", err)
				}
				contentParts = append(contentParts, openai.ChatMessagePart{
					Type: "audio_url",
					// sashabaranov/go-openai does not support audio input in their ChatCompletion API yet.
					// Substituting audio with image for now.
					// TODO: Once the support is there, fix this part of the code.
					ImageURL: &openai.ChatMessageImageURL{
						URL: fmt.Sprintf("data:audio/mpeg;base64,%s", base64Content),
					},
					// AudioURL: &openai.ChatMessageAudioURL{
					// 	URL: fmt.Sprintf("data:audio/mpeg;base64,%s", base64Content),
					// },
				})
			}
			chatMessage = openai.ChatCompletionMessage{
				Role:         msg.Role,
				MultiContent: contentParts,
			}
		} else {
			chatMessage = openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		openaiMessages = append(openaiMessages, chatMessage)
	}
	return openaiMessages, nil
}

// getChatCompletionRequest constructs an OpenAI ChatCompletionRequest from the model's settings and input messages.
func (model *OpenAIChat) getChatCompletionRequest(messages []openai.ChatCompletionMessage, stream bool) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:               model.Id,
		Messages:            messages,
		Temperature:         model.Temperature,
		TopP:                model.TopP,
		MaxCompletionTokens: model.MaxCompletionTokens,
		PresencePenalty:     model.PresencePenalty,
		FrequencyPenalty:    model.FrequencyPenalty,
		Stop:                model.Stop,
		LogProbs:            model.LogProbs,
		TopLogProbs:         model.TopLogProbs,
		N:                   model.N,
		User:                model.User,
		Stream:              stream,
	}
}

// ChatCompletion sends a synchronous chat request to OpenAI and returns the response.
// It converts input messages to OpenAI's format, makes the API call, and constructs a ModelResponse with usage data.
func (model *OpenAIChat) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	openaiMessages, err := convertMessageToOpenAIFormat(messages)
	if err != nil {
		utils.Logger.Error("Failed to convert messages", "error", err)
		return models.ModelResponse{}, fmt.Errorf("failed to convert messages: %w", err)
	}

	resp, err := model.client.CreateChatCompletion(ctx, model.getChatCompletionRequest(openaiMessages, false))
	if err != nil {
		utils.Logger.Error("Failed to get chat completion", "model", model.Id, "error", err)
		return models.ModelResponse{}, fmt.Errorf("failed to get chat completion for model %s: %w", model.Id, err)
	}

	if len(resp.Choices) == 0 {
		utils.Logger.Error("No response from model")
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
	openaiMessages, err := convertMessageToOpenAIFormat(messages)
	if err != nil {
		utils.Logger.Error("Failed to convert messages", "error", err)
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	stream, err := model.client.CreateChatCompletionStream(ctx, model.getChatCompletionRequest(openaiMessages, true))
	if err != nil {
		utils.Logger.Error("Failed to create stream", "error", err)
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
