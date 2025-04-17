package utils

import "testing"

func TestFirstNonEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "All empty",
			values:   []string{"", "", ""},
			expected: "",
		},
		{
			name:     "First non-empty",
			values:   []string{"first", "second", "third"},
			expected: "first",
		},
		{
			name:     "Second non-empty",
			values:   []string{"", "second", "third"},
			expected: "second",
		},
		{
			name:     "Last non-empty",
			values:   []string{"", "", "third"},
			expected: "third",
		},
		{
			name:     "Mixed empty and non-empty",
			values:   []string{"", "second", "", "fourth"},
			expected: "second",
		},
		{
			name:     "Single non-empty",
			values:   []string{"single"},
			expected: "single",
		},
		{
			name:     "Single empty",
			values:   []string{""},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FirstNonEmpty(tc.values...)
			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}
