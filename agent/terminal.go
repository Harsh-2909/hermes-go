package agent

import (
	"fmt"
	"strings"

	"github.com/Harsh-2909/hermes-go/tools"
	"github.com/Harsh-2909/hermes-go/utils"
)

// TerminalPrinter holds the state for rendering responses in the terminal
type TerminalPrinter struct {
	isMarkdown      bool             // Flag to indicate if the response is in Markdown format
	showUserMessage bool             // Flag to indicate if the user message should be shown
	termWidth       int              // Width of the terminal for formatting
	logs            string           // Logs to be displayed
	userMessage     string           // User message to be displayed
	toolCalls       []tools.ToolCall // List of tool calls made by the assistant
	response        string           // Response from the assistant
	errorMessage    string           // Error message to be displayed
	streamEnded     bool             // Flag to indicate if the streaming has ended
}

// buildContent constructs the final output string based on the render state
func (tp *TerminalPrinter) buildContent() string {
	/* Steps to build the content:
	1. Add logs if available
	2. Add user message if `showUserMessage` is true
	3. Add reasoning from tools or secondary models if available (for the future)
	4. Add thinking if available (for the future)
	5. Add tool calls if available
	6. Add response. Handle Markdown, word wrap, etc.
	7. Add citations if available (for the future)
	8. Add error if any at the end
	9. Return the output to be rendered by pterm.
	*/
	var output string
	var toolCallStr string

	// Logs
	if tp.logs != "" {
		output += tp.logs + "\n"
	}

	// User Message
	if tp.showUserMessage {
		output += utils.MessageBox(tp.userMessage, tp.termWidth)
	}

	// Tool Calls
	for _, toolCall := range tp.toolCalls {
		toolCallStr += fmt.Sprintf("• %s %s\n", toolCall.Name, toolCall.Arguments)
	}
	if toolCallStr != "" {
		toolCallStr = strings.TrimRight(toolCallStr, "\n")
		output += utils.ToolCallBox(toolCallStr, tp.termWidth)
	}

	// Response
	if tp.response != "" {
		if tp.isMarkdown {
			resp := utils.RenderMarkdown(tp.response, tp.termWidth)
			output += utils.ResponseBox(resp, tp.termWidth, false)
		} else {
			output += utils.ResponseBox(tp.response, tp.termWidth, true)
		}
	}

	// Error Message
	if tp.errorMessage != "" {
		output += utils.ErrorBox(tp.errorMessage, tp.termWidth)
	}
	return output
}
