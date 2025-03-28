package tools

import (
	"context"

	"github.com/Harsh-2909/hermes-go/utils"
)

// CalculatorTools is a toolkit that provides basic arithmetic operations.
type CalculatorTools struct {
	EnableAdd      bool // EnableAdd enables the Add tool
	EnableSubtract bool // EnableSubtract enables the Subtract tool
	EnableMultiply bool // EnableMultiply enables the Multiply tool
	EnableDivide   bool // EnableDivide enables the Divide tool
	EnableModulus  bool // EnableModulus enables the Modulus tool

	// EnableAll enables all tools in the toolkit.
	EnableAll bool
}

// Tools returns the list of tools in the toolkit.
// It sets up the tools based on the configuration of the toolkit.
// @return The list of tools in the toolkit
func (c *CalculatorTools) Tools() []Tool {
	tools := make([]Tool, 0)
	if c.EnableAdd || c.EnableAll {
		addTool, err := CreateToolFromMethod(c, "Add")
		if err == nil {
			tools = append(tools, addTool)
		} else {
			utils.Logger.Error("Failed to create Add tool", "error", err)
		}
	}
	if c.EnableSubtract || c.EnableAll {
		subtractTool, err := CreateToolFromMethod(c, "Subtract")
		if err == nil {
			tools = append(tools, subtractTool)
		} else {
			utils.Logger.Error("Failed to create Subtract tool", "error", err)
		}
	}
	if c.EnableMultiply || c.EnableAll {
		multiplyTool, err := CreateToolFromMethod(c, "Multiply")
		if err == nil {
			tools = append(tools, multiplyTool)
		} else {
			utils.Logger.Error("Failed to create Multiply tool", "error", err)
		}
	}
	if c.EnableDivide || c.EnableAll {
		divideTool, err := CreateToolFromMethod(c, "Divide")
		if err == nil {
			tools = append(tools, divideTool)
		} else {
			utils.Logger.Error("Failed to create Divide tool", "error", err)
		}
	}
	if c.EnableModulus || c.EnableAll {
		modulusTool, err := CreateToolFromMethod(c, "Modulus")
		if err == nil {
			tools = append(tools, modulusTool)
		} else {
			utils.Logger.Error("Failed to create Modulus tool", "error", err)
		}
	}
	return tools
}

// Add two numbers and return the result.
// @param a: The first number to add
// @param b: The second number to add
// @return The sum of the two numbers
func (c *CalculatorTools) Add(ctx context.Context, a, b float64) float64 {
	return a + b
}

// Subtract two numbers and return the result.
// @param a: The first number
// @param b: The second number
// @return The result of subtracting b from a
func (c *CalculatorTools) Subtract(ctx context.Context, a, b float64) float64 {
	return a - b
}

// Multiply two numbers and return the result.
// @param a: The first number
// @param b: The second number
// @return The result of multiplying a by b
func (c *CalculatorTools) Multiply(ctx context.Context, a, b float64) float64 {
	return a * b
}

// Divide two numbers and return the result.
// @param a: The first number
// @param b: The second number
// @return The result of dividing a by b. If b is 0, return 0.
func (c *CalculatorTools) Divide(ctx context.Context, a, b float64) float64 {
	if b == 0 {
		utils.Logger.Error("Attempt to divide by zero")
		return 0
	}
	return a / b
}

// Modulus two numbers and return the result.
// @param a: The first number
// @param b: The second number
// @return The result of modulus a by b. If b is 0, return 0.
func (c *CalculatorTools) Modulus(ctx context.Context, a, b int) int {
	if b == 0 {
		utils.Logger.Error("Attempt to modulus by zero")
		return 0
	}
	return a % b
}
