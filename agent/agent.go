// Package agent provides the core functionality for managing AI agent conversations.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/glamour"
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
	utils.Logger.Debug("System Message", "system_message", systemMessageContent)
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
			// fmt.Printf("\n%+v\n", response)
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
					result = fmt.Sprintf("Error: %v", err)
				}
				utils.Logger.Debug("Tool execution complete", "name", toolCall.Name, "result", result)
				agent.Messages = append(agent.Messages, models.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: toolCall.ID,
				})
			}

		} else if response.Event == "complete" {
			agent.Messages = append(agent.Messages, assistantMessage)
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

// printState holds the current state for streaming display
type printState struct {
	isThinking bool
	toolCalls  []string
	response   string
}

// buildContent constructs the display content based on the state
func buildContent(state printState, markdown bool) string {
	var content string
	if state.isThinking {
		content += "Thinking...\n\n"
	}
	for _, tc := range state.toolCalls {
		content += pterm.Yellow("Tool Call:") + " " + tc + "\n"
	}
	if state.response != "" {
		content += "Response:\n"
		if markdown {
			content += renderMarkdown(state.response)
		} else {
			content += state.response
		}
	}
	return content
}

// renderMarkdown renders text as markdown using glamour
func renderMarkdown(text string) string {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	rendered, err := renderer.Render(text)
	if err != nil {
		return text // Fallback to plain text if rendering fails
	}
	return rendered
}

// PrintResponse prints the agent's response with rich formatting
func (agent *Agent) PrintResponse(ctx context.Context, userMessage string, stream bool, showMessage bool, media ...models.Media) error {
	var userResponse string
	var finalResponse string

	state := printState{isThinking: true}
	area, err := pterm.DefaultArea.WithRemoveWhenDone(false).Start()
	if err != nil {
		pterm.Error.Println("Error:", err)
		return err
	}
	defer area.Stop()
	if showMessage {
		userResponse = utils.MessageBox(userMessage, true)
	}
	spinner, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).Start("Thinking...")
	defer spinner.Stop()
	if !stream {
		// Non streaming case
		if showMessage {
			area.Update(userResponse)
		}
		response, err := agent.Run(ctx, userMessage, media...)
		if err != nil {
			pterm.Error.Println("Error:", err)
			return err
		}
		agentResponse := response.Data
		if agent.Markdown {
			agentResponse = renderMarkdown(agentResponse)
		}
		if showMessage {
			finalResponse += userResponse
		}
		finalResponse += utils.ResponseBox(agentResponse, true)
		area.Update(finalResponse)
		spinner.Stop()

	} else {
		// Streaming case
		var agentResponse string
		markdown := agent.Markdown
		// area.Update(buildContent(state, markdown))
		if showMessage {
			area.Update(userResponse)
		}

		ch, err := agent.RunStream(ctx, userMessage, media...)
		if err != nil {
			pterm.Error.Println("Error:", err)
			return err
		}
		spinner.Stop()
		streamEnded := false
		for resp := range ch {
			finalResponse = ""
			switch resp.Event {
			case "chunk":
				agentResponse += resp.Data
				if agent.Markdown {
					agentResponse = renderMarkdown(agentResponse)
				}
				if showMessage {
					finalResponse += userResponse
				}
				finalResponse += utils.ResponseBox(agentResponse, true)
				area.Update(finalResponse)
				// state.isThinking = false
				// state.response += resp.Data
				// area.Update(buildContent(state, markdown))
			case "tool_call":
				if agent.ShowToolCalls {
					for _, tc := range resp.ToolCalls {
						toolCallStr := fmt.Sprintf("%s %s", tc.Name, tc.Arguments)
						state.toolCalls = append(state.toolCalls, toolCallStr)
					}
					area.Update(buildContent(state, markdown))
				}
			case "end":
				streamEnded = true

			case "error":
				// state.isThinking = false
				// area.Update(buildContent(state, markdown) + "\nError: " + resp.Data)
				streamEnded = true
			}
			if streamEnded {
				break
			}
		}
	}
	return nil
}
