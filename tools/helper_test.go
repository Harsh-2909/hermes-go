package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToolkit is a sample toolkit for testing CreateToolFromMethod.
type TestToolkit struct{}

// Add adds two integers and returns the result.
// @param a: The first integer
// @param b: The second integer
func (t *TestToolkit) Add(ctx context.Context, a, b int) int {
	return a + b
}

// Concat concatenates two strings and returns the result.
// @param s1: The first string
// @param s2: The second string
func (t *TestToolkit) Concat(ctx context.Context, s1, s2 string) string {
	return s1 + s2
}

// Divide divides two integers and returns the result or an error if divisor is zero.
// @param a: The dividend
// @param b: The divisor
func (t *TestToolkit) Divide(ctx context.Context, a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// Check returns "Yes" if the boolean is true, otherwise "No".
// @param flag: The boolean flag
func (t *TestToolkit) Check(ctx context.Context, flag bool) string {
	if flag {
		return "Yes"
	}
	return "No"
}

// NoCtx is a method without context.Context for error testing.
func (t *TestToolkit) NoCtx(a int) int {
	return a
}

// NoReturn is a method with no return values for error testing.
func (t *TestToolkit) NoReturn(ctx context.Context) {
}

// TooManyReturns is a method with too many return values for error testing.
func (t *TestToolkit) TooManyReturns(ctx context.Context) (int, string, error) {
	return 0, "", nil
}

// WithSlice is a method with an unsupported parameter type for error testing.
func (t *TestToolkit) WithSlice(ctx context.Context, s []int) int {
	return len(s)
}

func (t *TestToolkit) NoDoc(ctx context.Context, x int) int {
	// NoDoc is a method without documentation for error testing.
	// Notice how this comment is not a doc comment.
	return x
}

func TestCreateToolFromMethod(t *testing.T) {
	toolkit := &TestToolkit{}

	// Test successful creation and execution for Add method
	t.Run("Add", func(t *testing.T) {
		tool, err := CreateToolFromMethod(toolkit, "Add")
		assert.NoError(t, err)                                                             // Verify no error
		assert.Equal(t, "Add", tool.Name)                                                  // Verify Name
		assert.Equal(t, "Add adds two integers and returns the result.", tool.Description) // Verify Description

		// Verify Parameters
		expectedParams := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "integer",
					"description": "The first integer",
				},
				"b": map[string]interface{}{
					"type":        "integer",
					"description": "The second integer",
				},
			},
			"required": []string{"a", "b"},
		}
		expectedJSON, _ := json.Marshal(expectedParams)
		actualJSON, _ := json.Marshal(tool.Parameters)
		assert.JSONEq(t, string(expectedJSON), string(actualJSON))

		// Test Execute with valid inputs
		args := `{"a": 3, "b": 5}`
		result, err := tool.Execute(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "8", result) // int result is JSON-marshaled as "8"

		// Test Execute with missing parameter
		args = `{"a": 3}`
		_, err = tool.Execute(context.Background(), args)
		assert.Error(t, err) // Verify error for missing parameter 'b'

		// Test Execute with invalid type
		args = `{"a": "three", "b": 5}`
		_, err = tool.Execute(context.Background(), args)
		assert.Error(t, err) // Verify error for invalid type for parameter 'a'
	})

	// Test successful creation and execution for Concat method
	t.Run("Concat", func(t *testing.T) {
		tool, err := CreateToolFromMethod(toolkit, "Concat")
		assert.NoError(t, err)                                                                       // Verify no error
		assert.Equal(t, "Concat", tool.Name)                                                         // Verify Name
		assert.Equal(t, "Concat concatenates two strings and returns the result.", tool.Description) // Verify Description

		expectedParams := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"s1": map[string]interface{}{
					"type":        "string",
					"description": "The first string",
				},
				"s2": map[string]interface{}{
					"type":        "string",
					"description": "The second string",
				},
			},
			"required": []string{"s1", "s2"},
		}
		expectedJSON, _ := json.Marshal(expectedParams)
		actualJSON, _ := json.Marshal(tool.Parameters)
		assert.JSONEq(t, string(expectedJSON), string(actualJSON))

		args := `{"s1": "hello", "s2": "world"}`
		result, err := tool.Execute(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "\"helloworld\"", result) // string result is JSON-marshaled with quotes
	})

	// Test method with error return
	t.Run("Divide", func(t *testing.T) {
		tool, err := CreateToolFromMethod(toolkit, "Divide")
		assert.NoError(t, err) // Verify no error

		// Test with valid inputs
		args := `{"a": 10, "b": 2}`
		result, err := tool.Execute(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "5", result)

		// Test with division by zero
		args = `{"a": 10, "b": 0}`
		_, err = tool.Execute(context.Background(), args)
		assert.Error(t, err) // Verify error for division by zero
	})

	// Test method with boolean parameter
	t.Run("Check", func(t *testing.T) {
		tool, err := CreateToolFromMethod(toolkit, "Check")
		assert.NoError(t, err) // Verify no error

		// Test with true
		args := `{"flag": true}`
		result, err := tool.Execute(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "\"Yes\"", result)

		// Test with false
		args = `{"flag": false}`
		result, err = tool.Execute(context.Background(), args)
		assert.NoError(t, err)
		assert.Equal(t, "\"No\"", result)
	})

	// Error case: Non-existent method
	t.Run("NonExistentMethod", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "Subtract")
		assert.Error(t, err) // Verify error for non-existent method
	})

	// Error case: No context.Context parameter
	t.Run("NoCtxParam", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "NoCtx")
		assert.Error(t, err) // Verify error for method without context.Context
	})

	// Error case: No return values
	t.Run("NoReturn", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "NoReturn")
		assert.Error(t, err) // Verify error for method with no return values
	})

	// Error case: Too many return values
	t.Run("TooManyReturns", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "TooManyReturns")
		assert.Error(t, err) // Verify error for method with too many return values
	})

	// Error case: Unsupported parameter type
	t.Run("WithSlice", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "WithSlice")
		assert.Error(t, err) // Verify error for unsupported parameter type []int
	})

	// Error case: No doc comments
	t.Run("NoDoc", func(t *testing.T) {
		_, err := CreateToolFromMethod(toolkit, "NoDoc")
		assert.Error(t, err) // Verify error for method with no doc comments
	})
}
