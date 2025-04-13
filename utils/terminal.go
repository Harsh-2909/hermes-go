package utils

import (
	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
)

// PrintError prints an error message in red color.
func PrintError(err error) {
	pterm.Error.Println(err)
}

// PrintSuccess prints a success message in green color.
func PrintSuccess(msg string) {
	pterm.Success.Println(msg)
}

// PrintInfo prints an info message in blue color.
func PrintInfo(msg string) {
	pterm.Info.Println(msg)
}

// renderMarkdown renders text as markdown using glamour
func RenderMarkdown(text string, terminalWidth int) string {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(terminalWidth-2),
	)
	rendered, err := renderer.Render(text)
	if err != nil {
		return text // Fallback to plain text if rendering fails
	}
	return rendered
}

// PaddedBox is a custom box printer for hermes-go.
var PaddedBox = pterm.BoxPrinter{
	VerticalString:          "┃",
	TopRightCornerString:    "┗",
	TopLeftCornerString:     "┛",
	BottomLeftCornerString:  "┓",
	BottomRightCornerString: "┏",
	HorizontalString:        "━",
	BoxStyle:                &pterm.ThemeDefault.BoxStyle,
	TextStyle:               &pterm.ThemeDefault.BoxTextStyle,
	RightPadding:            1,
	LeftPadding:             1,
	TopPadding:              1,
	BottomPadding:           1,
	TitleTopLeft:            true,
}

// boxRenderer is a helper function that creates a box with the provided message.
// It handles all the common logic such as padding, text wrap, etc.
//
// It takes a box printer, message, and a boolean to extend the width of the box.
// It returns the modified box printer, with the wrapped message.
func boxRenderer(box *pterm.BoxPrinter, message string, termWidth int, wordWrap bool) (*pterm.BoxPrinter, string) {
	// Calculate available width for the text inside the box
	// Borders (2 characters: 1 left, 1 right) + Padding (2 characters: 1 left, 1 right)
	availableWidth := termWidth - 4
	messageLen := len(message)
	wrappedMessage := message
	if messageLen >= availableWidth {
		// Wrap the message if it exceeds the available width
		// Using pterm.DefaultParagraph to wrap the message
		if wordWrap {
			wrappedMessage = pterm.DefaultParagraph.WithMaxWidth(availableWidth).Sprint(message)
		}
		return box, wrappedMessage
	}

	// If the message fits within the available width, no wrapping is needed
	// Adjust the padding to make the box fit the whole terminal width
	box = box.WithRightPadding(availableWidth - messageLen)
	return box, wrappedMessage
}

// MessageBox creates a styled message box for the user's message
//
// The message box is styled with a cyan border and a light cyan title.
// The message is printed in green color.
func MessageBox(userMessage string, termWidth int) string {
	messageBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgCyan}).WithTitle(pterm.LightCyan("Message"))
	messageBox, wrappedMessage := boxRenderer(messageBox, userMessage, termWidth, true)
	return messageBox.Sprintfln(pterm.Green(wrappedMessage))
}

// ResponseBox creates a styled response box for the agent's response
//
// Note: In case of message already being wrapped, send wordWrap as false
//
// The response box is styled with a blue border and a light blue title.
// The message is printed in default color.
func ResponseBox(assistantMessage string, termWidth int, wordWrap bool) string {
	responseBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgBlue}).WithTitle(pterm.LightBlue("Response"))
	responseBox, wrappedMessage := boxRenderer(responseBox, assistantMessage, termWidth, wordWrap)
	return responseBox.Sprintfln(wrappedMessage)
}

// ThinkingBox creates a styled box for thinking states
//
// The thinking box is styled with a green border and a light green title.
// The message is printed in default color.
func ThinkingBox(message string, termWidth int) string {
	thinkingBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgGreen}).WithTitle(pterm.LightGreen("Thinking"))
	thinkingBox, wrappedMessage := boxRenderer(thinkingBox, message, termWidth, true)
	return thinkingBox.Sprintfln(wrappedMessage)
}

// ToolCallBox creates a styled box for tool call information
//
// The tool call box is styled with a yellow border and a light yellow title.
// The message is printed in yellow color.
func ToolCallBox(toolCall string, termWidth int) string {
	toolBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgYellow}).WithTitle(pterm.LightYellow("Tool Call"))
	toolBox, wrappedMessage := boxRenderer(toolBox, toolCall, termWidth, true)
	return toolBox.Sprintfln(pterm.Yellow(wrappedMessage))
}

// CitationBox creates a styled box for citations
//
// The citation box is styled with a gray border and a gray title.
// The message is printed in gray color.
func CitationBox(citation string, termWidth int) string {
	citationBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgGray}).WithTitle(pterm.Gray("Citation"))
	citationBox, wrappedMessage := boxRenderer(citationBox, citation, termWidth, true)
	return citationBox.Sprintfln(pterm.Gray(wrappedMessage))
}

// ErrorBox creates a styled box for error messages
//
// The error box is styled with a red border and a light red title.
// The message is printed in red color.
func ErrorBox(errMsg string, termWidth int) string {
	// Setting the style and title for the error box
	errorBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgRed}).WithTitle(pterm.LightRed("Error"))
	errorBox, wrappedMessage := boxRenderer(errorBox, errMsg, termWidth, true)
	return errorBox.Sprintfln(pterm.Red(wrappedMessage))
}
