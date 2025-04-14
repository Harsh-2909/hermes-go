package utils_test

import (
	"strings"
	"testing"

	"github.com/Harsh-2909/hermes-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestBoxFunctions(t *testing.T) {
	tests := []struct {
		name            string
		fn              func() string
		expectedTitle   string
		expectedMessage string
	}{
		{
			name: "MessageBox normal",
			fn: func() string {
				return utils.MessageBox("Hello, World!", 50)
			},
			expectedTitle:   "Message",
			expectedMessage: "Hello, World!",
		},
		{
			name: "MessageBox empty message",
			fn: func() string {
				return utils.MessageBox("", 50)
			},
			expectedTitle:   "Message",
			expectedMessage: "",
		},
		{
			name: "ResponseBox wrapping",
			fn: func() string {
				return utils.ResponseBox("Response Test", 30, true)
			},
			expectedTitle:   "Response",
			expectedMessage: "Response Test",
		},
		{
			name: "ThinkingBox long message",
			fn: func() string {
				longMsg := "This is a long thinking message that should be wrapped appropriately"
				return utils.ThinkingBox(longMsg, 40)
			},
			expectedTitle:   "Thinking",
			expectedMessage: "This is a long thinking message that should be wrapped appropriately",
		},
		{
			name: "ToolCallBox normal",
			fn: func() string {
				return utils.ToolCallBox("Tool call executed", 50)
			},
			expectedTitle:   "Tool Calls",
			expectedMessage: "Tool call executed",
		},
		{
			name: "CitationBox normal",
			fn: func() string {
				return utils.CitationBox("Citation information", 50)
			},
			expectedTitle:   "Citation",
			expectedMessage: "Citation information",
		},
		{
			name: "ErrorBox normal",
			fn: func() string {
				return utils.ErrorBox("Error occurred", 50)
			},
			expectedTitle:   "Error",
			expectedMessage: "Error occurred",
		},
		{
			name: "ResponseBox no word wrap",
			fn: func() string {
				return utils.ResponseBox("No wrap", 50, false)
			},
			expectedTitle:   "Response",
			expectedMessage: "No wrap",
		},
		{
			name: "MessageBox exact fit",
			fn: func() string {
				// Available width = termWidth - 4; for termWidth 20, message length = 16.
				return utils.MessageBox("1234567890123456", 20)
			},
			expectedTitle:   "Message",
			expectedMessage: "1234567890123456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.fn()
			// Check that the output contains the title and message.
			assert.Contains(t, output, tc.expectedTitle)
			// Strip ANSI codes for content check
			plainGot := utils.StripANSI(output)
			// Remove leading/trailing whitespace, newlines, and box characters
			ss := strings.Split(plainGot, "\n")
			for i := range ss {
				ss[i] = strings.Trim(ss[i], "┓")
				ss[i] = strings.Trim(ss[i], "┛")
				ss[i] = strings.Trim(ss[i], "┏")
				ss[i] = strings.Trim(ss[i], "┃")
				ss[i] = strings.Trim(ss[i], "━")
				ss[i] = strings.Trim(ss[i], "┗")
				ss[i] = strings.TrimSpace(ss[i])
			}
			plainGot = strings.Join(ss, " ")
			assert.Contains(t, plainGot, tc.expectedMessage)
			assert.True(t, strings.HasSuffix(output, "\n"))
		})
	}
}
