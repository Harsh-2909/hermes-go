// Package models provides implementations of the Model interface, including OpenAI integration.
package models

import (
	"context"
	"os"

	"github.com/Harsh-2909/hermes-go/models"
	openaiModel "github.com/Harsh-2909/hermes-go/models/openai"
	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/Harsh-2909/hermes-go/utils"
	"github.com/sashabaranov/go-openai"
)

const DeepSeekBaseURL = "https://api.deepseek.com"

// DeepSeek provides a struct for interacting with DeepSeek models.
//
// For more information, see: https://api-docs.deepseek.com/
type DeepSeek struct {
	ApiKey           string  // Required OpenAI API key. If not provided, it will be fetched from the environment variable `DEEPSEEK_API_KEY`.
	Id               string  // Required model ID (e.g., "deepseek-chat")
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

	baseChatModel openaiModel.BaseChat
}

// Init initializes the DeepSeek instance with defaults and validates required fields.
// It panics if ApiKey or Id is missing.
func (model *DeepSeek) Init() {
	if model.isInit {
		return
	}
	model.ApiKey = utils.FirstNonEmpty(model.ApiKey, os.Getenv("DEEPSEEK_API_KEY"))
	if model.ApiKey == "" {
		utils.Logger.Error("DeepSeek must have an API key")
		panic("DeepSeek must have an API key")
	}
	if model.Id == "" {
		utils.Logger.Error("DeepSeek must have a model ID")
		panic("DeepSeek must have a model ID")
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

	model.baseChatModel = openaiModel.BaseChat{
		BaseURL:             DeepSeekBaseURL,
		ApiKey:              model.ApiKey,
		Id:                  model.Id,
		Temperature:         model.Temperature,
		PresencePenalty:     model.PresencePenalty,
		FrequencyPenalty:    model.FrequencyPenalty,
		Stop:                model.Stop,
		N:                   model.N,
		User:                model.User,
		TopP:                model.TopP,
		MaxCompletionTokens: model.MaxCompletionTokens,
		LogProbs:            model.LogProbs,
		TopLogProbs:         model.TopLogProbs,

		Client: model.client,
	}
	model.baseChatModel.Init()
	model.isInit = true
}

func (model *DeepSeek) SetTools(tools []tools.Tool) {
	model.baseChatModel.SetTools(tools)
}

// ChatCompletion sends a synchronous chat request to OpenAI and returns the response.
// It converts input messages to OpenAI's format, makes the API call, and constructs a ModelResponse with usage data.
func (model *DeepSeek) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	return model.baseChatModel.ChatCompletion(ctx, messages)
}

// ChatCompletionStream initiates a streaming chat request to OpenAI and returns a channel of responses.
// It emits ModelResponse events ("chunk" for content, "end" for completion, "error" for failures).
// The caller must consume the channel to process the stream.
func (model *DeepSeek) ChatCompletionStream(ctx context.Context, messages []models.Message) (chan models.ModelResponse, error) {
	return model.baseChatModel.ChatCompletionStream(ctx, messages)
}
