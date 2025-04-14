package utils

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRenderMarkdown tests the RenderMarkdown function
func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		terminalWidth int
		wantOutput    string
		wantFallback  bool
	}{
		{
			name:          "valid markdown with headers and lists",
			input:         "# Header\n\n- Item 1\n- Item 2\n**Bold** text",
			terminalWidth: 80,
			wantOutput:    "\n  # Header                                                                  \n                                                                            \n  • Item 1                                                                  \n  • Item 2                                                                  \n  **Bold** text                                                             \n\n",
			wantFallback:  true,
		},
		{
			name:          "plain text input",
			input:         "Hello, world! This is plain text.",
			terminalWidth: 80,
			wantOutput:    "", // Expect plain text with possible wrapping
			wantFallback:  false,
		},
		{
			name:          "empty input",
			input:         "",
			terminalWidth: 80,
			wantOutput:    "\n\n", // Glamour always adds 2 newlines at the end of the output
			wantFallback:  true,
		},
		{
			name:          "invalid markdown with broken syntax",
			input:         "```broken code\nNo closing tag",
			terminalWidth: 80,
			wantOutput:    "\n                                                                            \n    No closing tag                                                          \n\n",
			wantFallback:  true, // Glamour will render it in markdown, it just won't be closed. This is important for us because this is how we want it to work (for streaming with markdown support)
		},
		{
			name:          "small terminal width",
			input:         "This is a very long sentence that should wrap because the terminal width is small.",
			terminalWidth: 20,
			wantOutput:    "\n  This is a very\n  long sentence \n  that should   \n  wrap because  \n  the terminal  \n  width is      \n  small.        \n\n", // Check wrapping
			wantFallback:  true,
		},
		{
			name:          "negative terminal width",
			input:         "Should handle negative width gracefully.",
			terminalWidth: -10,
			wantOutput:    "", // Expect rendering with default/fallback wrapping
			wantFallback:  false,
		},
		{
			name:          "unicode and special characters",
			input:         "Hello, 世界! 😊 Special chars: # * _",
			terminalWidth: 80,
			wantOutput:    "", // Expect proper handling of unicode
			wantFallback:  false,
		},
		{
			name:          "long paragraph",
			input:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			terminalWidth: 50,
			wantOutput:    "", // Check wrapping at 48 characters (50-2)
			wantFallback:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function
			got := RenderMarkdown(tt.input, tt.terminalWidth)

			if tt.wantFallback {
				// For fallback cases, output should match input
				assert.Equal(t, tt.wantOutput, got)
			} else {
				// For non-fallback cases, verify output is non-empty and contains input text
				if tt.input != "" {
					assert.NotEmpty(t, got, "Expected non-empty rendered output")
					// Strip ANSI codes for content check
					plainGot := stripANSI(got)
					// Remove leading/trailing whitespace and newlines
					ss := strings.Split(plainGot, "\n")
					for i := range ss {
						ss[i] = strings.TrimSpace(ss[i])
					}
					plainGot = strings.Join(ss, " ")
					assert.Contains(t, plainGot, strings.TrimSpace(tt.input), "Rendered output should contain input text")
				} else {
					assert.Empty(t, got, "Expected empty output for empty input")
				}

				// Verify word wrapping
				if tt.terminalWidth > 2 && tt.input != "" {
					lines := strings.Split(got, "\n")
					for _, line := range lines {
						// Account for ANSI codes inflating length; approximate check
						assert.LessOrEqual(t, len(stripANSI(line)), tt.terminalWidth, "Line exceeds word wrap width")
					}
				}
			}
		})
	}
}

// stripANSI removes ANSI escape codes for testing purposes
func stripANSI(s string) string {
	// Use regexp to remove common ANSI escape codes
	ansiRegexp := regexp.MustCompile(`\x1B\[[0-9;]*[a-zA-Z]`)
	s = ansiRegexp.ReplaceAllString(s, "")
	return s
}
