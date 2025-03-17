package tools

import (
	"context"
)

// Tool represents a single tool that the agent can call.
// TODO: Need to create a format for Parameters which will then be converted based on the model's requirements
type Tool struct {
	Name        string                                                 // Unique name of the tool
	Description string                                                 // Description for the model to understand the tool's purpose
	Parameters  map[string]interface{}                                 // JSON Schema for tool parameters
	Execute     func(ctx context.Context, args string) (string, error) // Function to execute the tool
}

// Tools returns a list of tools containing only the tool itself.
// This function exists to implement the ToolKit interface
func (t Tool) Tools() []Tool {
	return []Tool{t}
}

func NewTool(name, description string, parameters map[string]interface{}, execute func(ctx context.Context, args string) (string, error)) Tool {
	return Tool{
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Execute:     execute,
	}
}

// ToolCall represents a request from the model to call a tool.
type ToolCall struct {
	ID        string // Unique ID for the tool call (used in OpenAI's API)
	Name      string // Name of the tool to call
	Arguments string // JSON-encoded arguments for the tool
}

// ToolKit is an interface for structs that provide multiple tools.
type ToolKit interface {
	Tools() []Tool // Returns a list of tools provided by the struct
}
