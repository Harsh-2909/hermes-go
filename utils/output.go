package utils

import "github.com/charmbracelet/glamour"

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
