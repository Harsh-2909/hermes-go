// Package agent provides the core functionality for managing AI agent conversations.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/Harsh-2909/hermes-go/utils"
)

// Agent manages a conversation with an AI model, maintaining history and settings for the system message.
// It interacts with a provided Model to process user inputs and generate responses.
type Agent struct {
	Model    models.Model     // The AI model used for generating responses (e.g., OpenAIChat)
	Messages []models.Message // History of system, user, and assistant messages in the conversation

	// Settings for building the default system message

	SystemMessage string      // Custom system message; if set, overrides other settings
	Description   string      // Description of the agent, prepended to the default system message
	Goal          string      // Task goal, included in the system message within <your_goal> tags
	Instructions  interface{} // Instructions for the user, included in the system message within <instructions> tags. Can be a string or []string
	Role          string      // Agent's role, included in the system message within <your_role> tags
	Markdown      bool        // If true, instructs the model to format responses in Markdown

	// Agent Tools

	Tools         []tools.ToolKit // Tools are functions the model may generate JSON inputs for
	ShowToolCalls bool            // Show tool calls in Agent response

	// Logger related settings

	DebugMode bool // If true, enables debug mode for additional logging

	// Internal fields

	isInit bool         // Internal flag to track initialization
	_tools []tools.Tool // Internal list of tools. This is a flat list of tools from the ToolKits using `GetAllTools()`
}

// Init initializes the Agent with required settings and the system message.
// It panics if no Model is provided and ensures Messages is initialized before appending the system message.
func (agent *Agent) Init() {
	if agent.isInit {
		return
	}
	if agent.Model == nil {
		panic("Agent must have a model")
	}
	// Handles the logger initialization
	utils.InitLogger(agent.DebugMode)

	// Initialize the model
	agent.Model.Init()
	// Add tools to the model
	agent.addToolToModel()

	if agent.Messages == nil {
		agent.Messages = []models.Message{}
	}
	if len(agent.Messages) == 0 {
		systemMessage := agent.getSystemMessage()
		if systemMessage.Content != "" {
			agent.Messages = append(agent.Messages, systemMessage)
		}
	}
	agent.isInit = true
}

// GetAllTools returns all tools from the agent.
func (agent *Agent) GetAllTools() []tools.Tool {
	if len(agent._tools) > 0 {
		return agent._tools
	}
	agent._tools = agent.processTools()
	return agent._tools
}

// processTools processes the agent's tools and returns a flat list of tools.
func (agent *Agent) processTools() []tools.Tool {
	if len(agent.Tools) == 0 {
		return []tools.Tool{}
	}
	processedTools := make([]tools.Tool, 0)
	for _, tool := range agent.Tools {
		if t, ok := tool.(tools.Tool); ok {
			processedTools = append(processedTools, t)
		} else {
			processedTools = append(processedTools, tool.Tools()...)
		}
	}

	return processedTools
}

// addToolToModel adds the agent's tools to the model if any are provided.
// It processes the tools and sets them in the model.
func (agent *Agent) addToolToModel() {
	if len(agent.Tools) == 0 {
		return
	}
	utils.Logger.Debug("Adding tools to model")
	processedTools := agent.GetAllTools()
	for _, tool := range processedTools {
		utils.Logger.Debug("Tool added to model", "tool_name", tool.Name)
	}
	agent.Model.SetTools(processedTools)
	utils.Logger.Debug("Tools added to model")
}

// getSystemMessage constructs the initial system message based on the agent's settings.
// It uses SystemMessage if provided; otherwise, it builds a message from Description, Goal, Role,
// and adds Markdown instructions if enabled.
func (agent *Agent) getSystemMessage() models.Message {
	var systemMessageContent string
	if agent.SystemMessage != "" {
		systemMessageContent = agent.SystemMessage
	} else {
		// Build the default system message for the Agent.
		// First add the Agent description if provided
		if agent.Description != "" {
			systemMessageContent += agent.Description + "\n\n"
		}
		// Then add the Agent goal if provided
		if agent.Goal != "" {
			systemMessageContent += fmt.Sprintf("<your_goal>\n%s\n</your_goal>\n\n", agent.Goal)
		}
		// Then add the Agent role if provided
		if agent.Role != "" {
			systemMessageContent += fmt.Sprintf("<your_role>\n%s\n</your_role>\n\n", agent.Role)
		}
		// Then add instructions for the Agent
		if agent.Instructions != nil {
			switch instr := agent.Instructions.(type) {
			case string:
				if instr != "" { // Only include if non-empty
					systemMessageContent += fmt.Sprintf("<instructions>\n%s\n</instructions>\n\n", instr)
				}
			case []string:
				if len(instr) > 0 {
					systemMessageContent += "<instructions>"
					for _, instruction := range instr {
						systemMessageContent += fmt.Sprintf("\n- %s", instruction)
					}
					systemMessageContent += "\n</instructions>\n\n"
				}
			default:
				utils.Logger.Warn("Unsupported type for Instructions; expected string or []string")
			}
		}
	}
	var additionalInformation []string
	if agent.Markdown {
		additionalInformation = append(additionalInformation, "Use markdown to format your answers.")
	}
	// Add additional information if provided
	if len(additionalInformation) > 0 {
		systemMessageContent += "<additional_information>"
		for _, instruction := range additionalInformation {
			systemMessageContent += fmt.Sprintf("\n- %s", instruction)
		}
		systemMessageContent += "\n</additional_information>\n\n"
	}
	if systemMessageContent != "" {
		utils.Logger.Debug("============== system ==============")
		utils.Logger.Debug(systemMessageContent)
	}
	return models.Message{Role: "system", Content: systemMessageContent}
}

// AddMessage appends a new message with the specified role and content to the conversation history.
func (agent *Agent) AddMessage(role, content string, media []models.Media) {
	images := []*models.Image{}
	audio := []*models.Audio{}
	for _, m := range media {
		if m.GetType() == "image" {
			img := m.(*models.Image)
			images = append(images, img)
		}
		if m.GetType() == "audio" {
			aud := m.(*models.Audio)
			audio = append(audio, aud)
		}
	}
	agent.Messages = append(agent.Messages, models.Message{Role: role, Content: content, Images: images, Audios: audio})
}

func findTool(tools []tools.Tool, name string) (*tools.Tool, error) {
	for _, tool := range tools {
		if tool.Name == name {
			return &tool, nil
		}
	}
	return nil, fmt.Errorf("tool %s not found", name)
}

// Run processes a user message synchronously and returns the model's response.
// It adds the user message to the history, invokes ChatCompletion on the Model, appends the assistant’s response,
// and returns the result. Returns an error if the model fails or no messages exist.
func (agent *Agent) Run(ctx context.Context, userMessage string, media ...models.Media) (models.ModelResponse, error) {
	agent.Init() // Ensure the agent is initialized
	utils.Logger.Debug("Agent Run Start")
	agent.AddMessage("user", userMessage, media)

	if len(agent.Messages) == 0 {
		return models.ModelResponse{}, fmt.Errorf("no messages available for chat completion")
	}

	// Save all the tool calls made by the assistant here. This will be returned in response
	var toolCalls []tools.ToolCall

	for {
		response, err := agent.Model.ChatCompletion(ctx, agent.Messages)
		if err != nil {
			return models.ModelResponse{}, err
		}

		assistantMessage := models.Message{
			Role:    "assistant",
			Content: response.Data,
		}
		if response.Event == "tool_call" {
			assistantMessage.ToolCalls = response.ToolCalls
			agent.Messages = append(agent.Messages, assistantMessage)

			for _, toolCall := range response.ToolCalls {
				tool, err := findTool(agent.GetAllTools(), toolCall.Name)
				if err != nil {
					utils.Logger.Error("Tool not found", "name", toolCall.Name, "error", err)
					agent.Messages = append(agent.Messages, models.Message{
						Role:       "tool",
						Content:    fmt.Sprintf("Error: tool %s not found", toolCall.Name),
						ToolCallID: toolCall.ID,
					})
					continue
				}
				utils.Logger.Debug("Executing tool", "name", toolCall.Name)
				result, err := tool.Execute(ctx, toolCall.Arguments)
				if err != nil {
					utils.Logger.Error("Tool execution failed", "name", toolCall.Name, "error", err)
					result = fmt.Sprintf("Error: %s", err.Error())
				}
				utils.Logger.Debug("Tool execution complete", "name", toolCall.Name, "result", result)
				agent.Messages = append(agent.Messages, models.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: toolCall.ID,
				})
				toolCalls = append(toolCalls, toolCall)
			}

		} else if response.Event == "complete" {
			agent.Messages = append(agent.Messages, assistantMessage)
			response.ToolCalls = toolCalls
			utils.Logger.Debug("Agent Run End")
			return response, nil
		} else {
			return models.ModelResponse{}, fmt.Errorf("unexpected event type: %s", response.Event)
		}
	}
}

// RunStream processes a user message and returns a channel for streaming model responses.
// It adds the user message to the history and invokes ChatCompletionStream on the Model.
// The caller must consume the channel to receive response chunks; the history is not updated here
// due to the streaming nature (see implementation note).
func (agent *Agent) RunStream(ctx context.Context, userMessage string, media ...models.Media) (chan models.ModelResponse, error) {
	agent.Init() // Ensure the agent is initialized
	utils.Logger.Debug("Agent RunStream Start")
	agent.AddMessage("user", userMessage, media)

	if len(agent.Messages) == 0 {
		return nil, fmt.Errorf("no messages available for chat completion")
	}

	// Accumulate response in the background for history.
	// TODO: Look into a better way to handle this, as it may not be ideal for large responses.
	ch := make(chan models.ModelResponse)
	go func() {
		defer close(ch)
		for {
			respCh, err := agent.Model.ChatCompletionStream(ctx, agent.Messages)
			if err != nil {
				ch <- models.ModelResponse{
					Event:     "error",
					Data:      err.Error(),
					CreatedAt: time.Now(),
				}
				return
			}

			fullResponse := ""
			var toolCalls []tools.ToolCall
			for resp := range respCh {
				if resp.Event == "chunk" {
					fullResponse += resp.Data
					ch <- resp // Forward content to the user
				} else if resp.Event == "tool_call" {
					toolCalls = resp.ToolCalls
					if resp.Data != "" {
						fullResponse += resp.Data
						ch <- models.ModelResponse{
							Event:     "chunk",
							Data:      resp.Data,
							CreatedAt: time.Now(),
						}
					}
					// Send a separate event for tool calls
					ch <- models.ModelResponse{
						Event:     "tool_call",
						ToolCalls: resp.ToolCalls,
						CreatedAt: time.Now(),
					}
				} else if resp.Event == "end" {
					// Break from the loop and handle the logic outside the response channel loop
					break
				}
			}
			assistantMessage := models.Message{
				Role:    "assistant",
				Content: fullResponse,
			}

			if len(toolCalls) > 0 {
				// Add assistant message with tool call
				assistantMessage.ToolCalls = toolCalls
				agent.Messages = append(agent.Messages, assistantMessage)

				// Execute tools and add results in Messages
				for _, toolCall := range toolCalls {
					tool, err := findTool(agent.GetAllTools(), toolCall.Name)
					if err != nil {
						utils.Logger.Error("Tool not found", "name", toolCall.Name, "error", err)
						agent.Messages = append(agent.Messages, models.Message{
							Role:       "tool",
							Content:    fmt.Sprintf("Error: tool %s not found", toolCall.Name),
							ToolCallID: toolCall.ID,
						})
						continue
					}
					utils.Logger.Debug("Executing tool", "name", toolCall.Name)
					result, err := tool.Execute(ctx, toolCall.Arguments)
					if err != nil {
						utils.Logger.Error("Tool execution failed", "name", toolCall.Name, "error", err)
						result = fmt.Sprintf("Error: %v", err)
					}
					utils.Logger.Debug("Tool execution complete", "name", toolCall.Name, "result", result)
					agent.Messages = append(agent.Messages, models.Message{
						Role:       "tool",
						Content:    result,
						ToolCallID: toolCall.ID,
					})
				}
			} else {
				// Add assistant message without tool call
				agent.Messages = append(agent.Messages, assistantMessage)
				// Send the end event to the channel
				ch <- models.ModelResponse{
					Event:     "end",
					CreatedAt: time.Now(),
				}
				break
			}
		}
		utils.Logger.Debug("Agent RunStream End")
	}()
	return ch, nil
}

// renderState holds the state for rendering responses
type renderState struct {
	termWidth    int
	isMarkdown   bool
	userMessage  string
	response     string
	toolCalls    []tools.ToolCall
	errorMessage string
	streamEnded  bool
}

// buildContent constructs the final output string based on the render state
func buildContent(state renderState, showMessage bool) string {
	/* Steps to build the content:
	1. Add user message if `showMessage` is true
	2. Add reasoning from tools or secondary models if available (for the future)
	3. Add thinking if available (for the future)
	4. Add tool calls if available
	5. Add response. Handle Markdown, word wrap, etc.
	6. Add citations if available (for the future)
	7. Add error if any at the end
	8. Return the output to be rendered by pterm.
	*/
	var output string
	var toolCallStr string

	// User Message
	if showMessage {
		output += utils.MessageBox(state.userMessage, state.termWidth)
	}

	// Tool Calls
	for _, toolCall := range state.toolCalls {
		toolCallStr += fmt.Sprintf("• %s %s\n", toolCall.Name, toolCall.Arguments)
	}
	if toolCallStr != "" {
		toolCallStr = strings.TrimRight(toolCallStr, "\n")
		output += utils.ToolCallBox(toolCallStr, state.termWidth)
	}

	// Response
	if state.response != "" {
		if state.isMarkdown {
			resp := utils.RenderMarkdown(state.response, state.termWidth)
			output += utils.ResponseBox(resp, state.termWidth, false)
		} else {
			output += utils.ResponseBox(state.response, state.termWidth, true)
		}
	}

	// Error Message
	if state.errorMessage != "" {
		output += utils.ErrorBox(state.errorMessage, state.termWidth)
	}
	return output
}

// PrintResponse prints the agent's response with rich formatting
func (agent *Agent) PrintResponse(ctx context.Context, userMessage string, stream bool, showMessage bool, media ...models.Media) error {
	agent.Init() // Ensure the agent is initialized
	// Fetch terminal width once at the start
	termWidth, _, err := pterm.GetTerminalSize()
	if err != nil {
		termWidth = 100 // Fallback to default width
	}

	// Initialize render state
	state := renderState{
		termWidth:   termWidth,
		userMessage: userMessage,
		isMarkdown:  agent.Markdown,
	}

	// Set up the area for rendering
	area, err := pterm.DefaultArea.WithRemoveWhenDone(false).Start()
	if err != nil {
		utils.Logger.Error("Unexpected error", "error", err)
		return err
	}
	defer area.Stop()

	// Show spinner while thinking
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).Start("Thinking...")
	defer spinner.Stop()

	if !stream {
		// Non-streaming case
		if showMessage {
			area.Update(utils.MessageBox(state.userMessage, state.termWidth))
		}
		response, err := agent.Run(ctx, userMessage, media...)
		if err != nil {
			state.errorMessage = err.Error()
		}
		spinner.Stop()
		state.toolCalls = response.ToolCalls
		state.response = response.Data
		area.Update(buildContent(state, showMessage))
	} else {
		// Streaming case
		if showMessage {
			area.Update(utils.MessageBox(state.userMessage, state.termWidth))
		}
		ch, err := agent.RunStream(ctx, userMessage, media...)
		if err != nil {
			state.errorMessage = err.Error()
		}
		spinner.Stop()
		for resp := range ch {
			switch resp.Event {
			case "chunk":
				state.response += resp.Data
				area.Update(buildContent(state, showMessage))
			case "tool_call":
				state.toolCalls = append(state.toolCalls, resp.ToolCalls...)
				area.Update(buildContent(state, showMessage))
			case "end":
				state.streamEnded = true
			case "error":
				state.errorMessage = resp.Data
				area.Update(buildContent(state, showMessage))
				state.streamEnded = true
			}
			if state.streamEnded {
				break
			}
		}
	}
	return nil
}
