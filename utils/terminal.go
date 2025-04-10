package utils

import (
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

func boxRenderer(box *pterm.BoxPrinter, message string, extendWidth bool) (*pterm.BoxPrinter, string) {
	var wrappedMessage string
	messageLen := len(message)
	w := 100 // Setting default terminal width to 100

	// Get the terminal size to adjust the box padding
	if extendWidth {
		width, _, err := pterm.GetTerminalSize()
		if err != nil {
			width = 100 // Fallback to default width if error occurs
		}
		w = width
	}

	// Calculate available width for the text inside the box
	// Borders (2 characters: 1 left, 1 right) + Padding (2 characters: 1 left, 1 right)
	availableWidth := w - 4
	if messageLen >= availableWidth {
		// Wrap the message if it exceeds the available width
		// Using pterm.DefaultParagraph to wrap the message
		wrappedMessage = pterm.DefaultParagraph.WithMaxWidth(availableWidth).Sprint(message)
	} else {
		// If the message fits within the available width, no wrapping is needed
		// Adjust the padding to make the box fit the whole terminal width
		wrappedMessage = message
		box = box.WithRightPadding(availableWidth - messageLen)
	}
	return box, wrappedMessage
}

// MessageBox creates a message box with the provided user message and returns the formatted string.
// It adjusts the padding based on the terminal size.
// The message box is styled with a cyan border and a light cyan title.
// The message is printed in green color.
func MessageBox(userMessage string, extendWidth bool) string {
	// Setting the style and title for the message box
	messageBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgCyan}).WithTitle(pterm.LightCyan("Message"))
	messageBox, wrappedMessage := boxRenderer(messageBox, userMessage, extendWidth)
	return messageBox.Sprintfln(pterm.Green(wrappedMessage))
}

// ResponseBox creates a response box with the provided assistant message and returns the formatted string.
// It adjusts the padding based on the terminal size.
// The response box is styled with a blue border and a light blue title.
// The message is printed in default color.
func ResponseBox(assistantMessage string, extendWidth bool) string {
	// Setting the style and title for the response box
	responseBox := PaddedBox.WithBoxStyle(&pterm.Style{pterm.FgBlue}).WithTitle(pterm.LightBlue("Response"))
	responseBox, wrappedMessage := boxRenderer(responseBox, assistantMessage, extendWidth)
	return responseBox.Sprintfln(wrappedMessage)
}
