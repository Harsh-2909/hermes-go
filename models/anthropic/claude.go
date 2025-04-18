// Package models provides implementations of the Model interface, including Anthropic integration.
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/Harsh-2909/hermes-go/utils"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Claude provides a struct for interacting with Anthropic Claude models.
//
// For more information, see: https://docs.anthropic.com/en/api/messages
type Claude struct {
	ApiKey      string  // Required Anthropic API key. If not provided, it will be fetched from the environment variable `ANTHROPIC_API_KEY`.
	Id          string  // Required model ID (e.g., "claude-3-sonnet-20240229")
	Temperature float32 // In [0,1] range. Higher values -> more creative
	TopP        float32 // Nucleus sampling parameter, in [0,1] range
	MaxTokens   int     // Maximum tokens to generate (required by Anthropic)

	// Internal fields
	client *anthropic.Client // Internal Anthropic API client
	isInit bool              // Tracks initialization
	tools  []tools.Tool      // List of tools for the model
}

// Init initializes the Claude instance, validating required fields and setting up the client.
// It panics if ApiKey or Id is missing.
func (model *Claude) Init() {
	if model.isInit {
		return
	}
	model.ApiKey = utils.FirstNonEmpty(model.ApiKey, os.Getenv("ANTHROPIC_API_KEY"))
	if model.ApiKey == "" {
		utils.Logger.Error("Claude must have an API key")
		panic("Claude must have an API key")
	}
	if model.Id == "" {
		utils.Logger.Error("Claude must have a model ID")
		panic("Claude must have a model ID")
	}
	if model.Temperature < 0 || model.Temperature > 1 {
		model.Temperature = 1.0 // Anthropic default
	}
	if model.TopP < 0 || model.TopP > 1 {
		model.TopP = 1.0 // Anthropic default
	}
	if model.MaxTokens <= 0 {
		model.MaxTokens = 4096 // Anthropic requires this; default to a reasonable value
	}

	if model.client == nil {
		// Initialize the client with the provided API key
		client := anthropic.NewClient(option.WithAPIKey(model.ApiKey))
		model.client = &client
	}
	model.isInit = true
}

// SetTools stores the provided tools in the model for use in API requests.
func (model *Claude) SetTools(tools []tools.Tool) {
	model.tools = tools
}

// formatMessages converts framework Messages to Anthropic's message format.
// It handles text, images, tool calls, and tool results, grouping tool results into subsequent user messages.
func formatMessages(messages []models.Message) ([]anthropic.MessageParam, string, error) {
	var anthropicMessages []anthropic.MessageParam
	var systemMessages []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			// System messages are added to the systemMessages slice
			if msg.Content != "" {
				systemMessages = append(systemMessages, msg.Content)
			}
		case "user":
			content := []anthropic.ContentBlockParamUnion{}

			// Add text content if present
			if msg.Content != "" {
				content = append(content, anthropic.ContentBlockParamUnion{
					OfRequestTextBlock: &anthropic.TextBlockParam{Text: msg.Content},
				})
			}

			// Add images if present
			for _, img := range msg.Images {
				// If URL is provided, use it directly. No need to encode to Base64
				if img.URL != "" {
					content = append(content, anthropic.ContentBlockParamUnion{
						OfRequestImageBlock: &anthropic.ImageBlockParam{
							Type: "image",
							Source: anthropic.ImageBlockParamSourceUnion{
								OfURLImageSource: &anthropic.URLImageSourceParam{
									URL: img.URL,
								},
							},
						},
					})
					continue
				}
				base64Content, err := img.Content()
				if err != nil {
					utils.Logger.Error("failed to get image content", "error", err)
					continue
				}
				mediaType, err := img.GetMediaType()
				if err != nil {
					utils.Logger.Error("failed to get media type for image", "error", err)
					continue
				}
				content = append(content, anthropic.NewImageBlockBase64(mediaType, base64Content))
			}

			// Audio not supported by Anthropic yet; ignore for now
			if len(msg.Audios) > 0 {
				utils.Logger.Warn("Audio inputs are not supported by Anthropic API; ignoring")
			}
			anthropicMessages = append(anthropicMessages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleUser,
				Content: content,
			})
		case "assistant":
			content := []anthropic.ContentBlockParamUnion{}

			// Add text content if present
			if msg.Content != "" {
				content = append(content, anthropic.ContentBlockParamUnion{
					OfRequestTextBlock: &anthropic.TextBlockParam{Text: msg.Content},
				})
			}

			// Add the tool calls initiated by the model to the message history
			for _, tc := range msg.ToolCalls {
				// Anthropic expects the `Input` field to be a JSON data instead of json string.
				// Thus, we need to unmarshal the `ToolCall.Arguments`
				var inputJSON map[string]any
				if err := json.Unmarshal([]byte(tc.Arguments), &inputJSON); err != nil {
					utils.Logger.Error("failed to parse tool call arguments", "error", err)
					continue
				}
				content = append(content, anthropic.ContentBlockParamUnion{
					OfRequestToolUseBlock: &anthropic.ToolUseBlockParam{
						ID:    tc.ID,
						Name:  tc.Name,
						Input: inputJSON,
					},
				})
			}
			anthropicMessages = append(anthropicMessages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleAssistant,
				Content: content,
			})
		case "tool":
			var content []anthropic.ContentBlockParamUnion
			content = append(content, anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false))
			anthropicMessages = append(anthropicMessages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleUser,
				Content: content,
			})
		default:
			utils.Logger.Error("unsupported message role", "role", msg.Role)
		}
	}
	return anthropicMessages, strings.Join(systemMessages, "\n"), nil
}

// getChatCompletionRequest constructs a ChatCompletionRequest from the model's settings and input messages.
func (model *Claude) getChatCompletionRequest(messages []anthropic.MessageParam, systemMessage string) anthropic.MessageNewParams {
	// Convert tools to Anthropic format
	var anthropicTools []anthropic.ToolUnionParam
	for _, tool := range model.tools {
		tool := anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: tool.Parameters["properties"],
			},
		}
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{OfTool: &tool})
	}

	chatCompletionRequest := anthropic.MessageNewParams{
		Model:       model.Id,
		MaxTokens:   int64(model.MaxTokens),
		Temperature: anthropic.Float(float64(model.Temperature)),
		TopP:        anthropic.Float(float64(model.TopP)),
		Messages:    messages,
		Tools:       anthropicTools,
	}

	// Set system message only if provided
	if systemMessage != "" {
		chatCompletionRequest.System = []anthropic.TextBlockParam{
			{Text: systemMessage},
		}
	}
	return chatCompletionRequest
}

// ChatCompletion sends a synchronous chat request to Anthropic and returns the response.
func (model *Claude) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	anthropicMessages, systemMessage, err := formatMessages(messages)
	// DEBUG: Check messages going to Anthropic API
	// fmt.Printf("\n\nClaude Messages:\n")
	// for _, msg := range anthropicMessages {
	// 	val, _ := msg.MarshalJSON()
	// 	fmt.Printf("%s\n", string(val))
	// }
	if err != nil {
		utils.Logger.Error("Failed to convert messages", "error", err)
		return models.ModelResponse{}, fmt.Errorf("failed to convert messages: %w", err)
	}

	resp, err := model.client.Messages.New(ctx, model.getChatCompletionRequest(anthropicMessages, systemMessage))
	if err != nil {
		utils.Logger.Error("Failed to get chat completion", "model", model.Id, "error", err)
		return models.ModelResponse{}, fmt.Errorf("failed to get chat completion for model %s: %w", model.Id, err)
	}

	if len(resp.Content) == 0 {
		utils.Logger.Error("No response from model")
		return models.ModelResponse{}, fmt.Errorf("no response from model")
	}

	modelResp := models.ModelResponse{
		CreatedAt: time.Now(),
	}
	// fmt.Printf("\nMessage Received: %s\n", resp.RawJSON()) // DEBUG: Check messages received from Anthropic API

	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			modelResp.Data += variant.Text
		case anthropic.ToolUseBlock:
			modelResp.Event = "tool_call"
			modelResp.ToolCalls = append(modelResp.ToolCalls, tools.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		case anthropic.ThinkingBlock:
			utils.Logger.Warn("thinking block not supported yet", "thinking", variant.Thinking)
		case anthropic.RedactedThinkingBlock:
			utils.Logger.Warn("redacted thinking block encountered", "data", variant.Data)
		default:
			utils.Logger.Error("unknown block type", "block", variant)
			return models.ModelResponse{}, fmt.Errorf("unknown block type: %T", variant)
		}
	}

	if resp.StopReason == "tool_use" {
		modelResp.Event = "tool_call"
	} else {
		modelResp.Event = "complete"
	}

	// Usage data
	modelResp.Usage = &models.Usage{
		PromptTokens:     int(resp.Usage.InputTokens),
		CompletionTokens: int(resp.Usage.OutputTokens),
		TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
	}

	return modelResp, nil
}

// ChatCompletionStream initiates a streaming chat request to Anthropic and returns a channel of responses.
// It emits "chunk" for content, "tool_call" for tool use, "end" for completion, or "error" for failures.
func (model *Claude) ChatCompletionStream(ctx context.Context, messages []models.Message) (chan models.ModelResponse, error) {
	anthropicMessages, systemMessage, err := formatMessages(messages)
	// DEBUG: Check messages going to Anthropic API
	// fmt.Printf("\n\nClaude Messages:\n")
	// for _, msg := range anthropicMessages {
	// 	val, _ := msg.MarshalJSON()
	// 	fmt.Printf("%s\n", string(val))
	// }
	if err != nil {
		utils.Logger.Error("Failed to convert messages", "error", err)
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	stream := model.client.Messages.NewStreaming(ctx, model.getChatCompletionRequest(anthropicMessages, systemMessage))
	ch := make(chan models.ModelResponse)
	go func() {
		defer close(ch)
		defer stream.Close()

		content := ""
		toolCalls := make(map[int]*tools.ToolCall)
		message := anthropic.Message{}

		for stream.Next() {
			event := stream.Current()
			err := message.Accumulate(event)
			if err != nil {
				ch <- models.ModelResponse{
					Event:     "error",
					Data:      err.Error(),
					CreatedAt: time.Now(),
				}
				return
			}
			// fmt.Printf("Event received: %s\n", event.RawJSON()) // DEBUG: Check messages received from Anthropic API

			// Check Anthropic docs on streaming:
			// https://docs.anthropic.com/en/api/messages-streaming
			switch variant := event.AsAny().(type) {
			case anthropic.MessageStartEvent:
			case anthropic.MessageDeltaEvent:
			case anthropic.MessageStopEvent:
				// This is the last event of the stream
				// Break out of for loop, end event will be sent after leaving the loop
				break

			case anthropic.ContentBlockStartEvent:
				switch block := variant.ContentBlock.AsAny().(type) {
				case anthropic.TextBlock:
					content += block.Text
					ch <- models.ModelResponse{
						Event:     "chunk",
						Data:      block.Text,
						CreatedAt: time.Now(),
					}
				case anthropic.ToolUseBlock:
					toolCalls[int(variant.Index)] = &tools.ToolCall{
						ID:   block.ID,
						Name: block.Name,
					}
				case anthropic.ThinkingBlock:
				case anthropic.RedactedThinkingBlock:
				default:
					utils.Logger.Error("unknown content block type", "block", block)
				}

			case anthropic.ContentBlockDeltaEvent:
				switch block := variant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					content += block.Text
					ch <- models.ModelResponse{
						Event:     "chunk",
						Data:      block.Text,
						CreatedAt: time.Now(),
					}
				case anthropic.InputJSONDelta:
					if tc, exists := toolCalls[int(variant.Index)]; exists {
						tc.Arguments += block.PartialJSON
					}
				case anthropic.CitationsDelta:
				case anthropic.ThinkingDelta:
				case anthropic.SignatureDelta:
				default:
					utils.Logger.Error("unknown content block type", "block", block)
				}

			case anthropic.ContentBlockStopEvent:
			}
		}

		// Check for any errors in the stream
		if stream.Err() != nil {
			ch <- models.ModelResponse{
				Event:     "error",
				Data:      stream.Err().Error(),
				CreatedAt: time.Now(),
			}
			return
		}

		// After streaming ends, check for tool calls
		if len(toolCalls) > 0 {
			var finalToolCalls []tools.ToolCall
			for _, tc := range toolCalls {
				finalToolCalls = append(finalToolCalls, *tc)
			}
			ch <- models.ModelResponse{
				Event:     "tool_call",
				ToolCalls: finalToolCalls,
				CreatedAt: time.Now(),
			}
		}

		// Send the final message after tool calls
		ch <- models.ModelResponse{
			Event:     "end",
			CreatedAt: time.Now(),
			Usage: &models.Usage{
				PromptTokens:     int(message.Usage.InputTokens),
				CompletionTokens: int(message.Usage.OutputTokens),
				TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
			},
		}
	}()

	return ch, nil
}
