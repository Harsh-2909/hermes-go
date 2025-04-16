package tools

import (
	"context"
	"math"

	"github.com/Harsh-2909/hermes-go/utils"
)

// CalculatorTools is a toolkit that provides basic arithmetic operations.
type CalculatorTools struct {
	EnableAdd          bool // EnableAdd enables the Add tool
	EnableSubtract     bool // EnableSubtract enables the Subtract tool
	EnableMultiply     bool // EnableMultiply enables the Multiply tool
	EnableDivide       bool // EnableDivide enables the Divide tool
	EnableModulus      bool // EnableModulus enables the Modulus tool
	EnableExponentiate bool // EnableExponentiate enables the Exponentiate tool
	EnableFactorial    bool // EnableFactorial enables the Factorial tool
	EnableIsPrime      bool // EnableIsPrime enables the IsPrime tool
	EnableSquareRoot   bool // EnableSquareRoot enables the SquareRoot tool

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
	if c.EnableExponentiate || c.EnableAll {
		expTool, err := CreateToolFromMethod(c, "Exponentiate")
		if err == nil {
			tools = append(tools, expTool)
		} else {
			utils.Logger.Error("Failed to create Exponentiate tool", "error", err)
		}
	}
	if c.EnableFactorial || c.EnableAll {
		factTool, err := CreateToolFromMethod(c, "Factorial")
		if err == nil {
			tools = append(tools, factTool)
		} else {
			utils.Logger.Error("Failed to create Factorial tool", "error", err)
		}
	}
	if c.EnableIsPrime || c.EnableAll {
		primeTool, err := CreateToolFromMethod(c, "IsPrime")
		if err == nil {
			tools = append(tools, primeTool)
		} else {
			utils.Logger.Error("Failed to create IsPrime tool", "error", err)
		}
	}
	if c.EnableSquareRoot || c.EnableAll {
		sqrtTool, err := CreateToolFromMethod(c, "SquareRoot")
		if err == nil {
			tools = append(tools, sqrtTool)
		} else {
			utils.Logger.Error("Failed to create SquareRoot tool", "error", err)
		}
	}
	return tools
}

// Add two numbers and return the result.
//
// @param a: The first number to add
// @param b: The second number to add
// @return The sum of the two numbers
func (c *CalculatorTools) Add(ctx context.Context, a, b float64) float64 {
	return a + b
}

// Subtract two numbers and return the result.
//
// @param a: The first number
// @param b: The second number
// @return The result of subtracting b from a
func (c *CalculatorTools) Subtract(ctx context.Context, a, b float64) float64 {
	return a - b
}

// Multiply two numbers and return the result.
//
// @param a: The first number
// @param b: The second number
// @return The result of multiplying a by b
func (c *CalculatorTools) Multiply(ctx context.Context, a, b float64) float64 {
	return a * b
}

// Divide two numbers and return the result.
//
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
//
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

// Exponentiate returns base raised to the power exp.
// @param base: The base number
// @param exp: The exponent
// @return The result of base^exp
func (c *CalculatorTools) Exponentiate(ctx context.Context, base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Factorial returns the factorial of n.
// @param n: The number to compute factorial for
// @return The factorial of n. For n < 0, returns 0.
func (c *CalculatorTools) Factorial(ctx context.Context, n int) int {
	if n < 0 {
		utils.Logger.Error("Factorial of negative number", "n", n)
		return 0
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

// IsPrime determines if n is a prime number.
// @param n: The number to check
// @return True if n is prime, otherwise false.
func (c *CalculatorTools) IsPrime(ctx context.Context, n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// SquareRoot returns the square root of x.
// @param x: The number to find the square root of
// @return The square root of x. If x is negative, returns 0.
func (c *CalculatorTools) SquareRoot(ctx context.Context, x float64) float64 {
	if x < 0 {
		utils.Logger.Error("Square root of negative number", "x", x)
		return 0
	}
	return math.Sqrt(x)
}
