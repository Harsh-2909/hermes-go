package tools

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatorTools_Tools(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableAll: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 9, len(tools))
	assert.NotNil(t, tools)
}

func TestCalculatorTools_Tools_Add(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableAdd: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	assert.NotNil(t, tools)

	tool := tools[0]
	ctx := context.Background()
	val, err := tool.Execute(ctx, `{"a": 2, "b": 3}`)
	assert.Equal(t, "Add", tool.Name)
	assert.Equal(t, "Add two numbers and return the result.", tool.Description)
	assert.Equal(t, "5", val)
	assert.Nil(t, err)
	assert.Equal(t, 5.0, calcTools.Add(ctx, 2, 3))
}

func TestCalculatorTools_Tools_Subtract(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableSubtract: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	assert.NotNil(t, tools)

	tool := tools[0]
	ctx := context.Background()
	val, err := tool.Execute(ctx, `{"a": 2, "b": 3}`)
	assert.Equal(t, "Subtract", tool.Name)
	assert.Equal(t, "Subtract two numbers and return the result.", tool.Description)
	assert.Equal(t, "-1", val)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, calcTools.Subtract(ctx, 2, 3))
}

func TestCalculatorTools_Tools_Multiply(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableMultiply: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	assert.NotNil(t, tools)

	tool := tools[0]
	ctx := context.Background()
	val, err := tool.Execute(ctx, `{"a": 4, "b": 3}`)
	assert.Equal(t, "Multiply", tool.Name)
	assert.Equal(t, "Multiply two numbers and return the result.", tool.Description)
	assert.Equal(t, "12", val)
	assert.Nil(t, err)
	assert.Equal(t, 6.0, calcTools.Multiply(ctx, 2, 3))
}

func TestCalculatorTools_Tools_Divide(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableDivide: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	assert.NotNil(t, tools)

	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		test   string
		a, b   float64
		result float64
	}{
		{"Floating point division", 5, 2, 2.50},
		{"Recurring division", 4, 3, 1.3333333333333333},
		{"Zero reminder division", 6, 3, 2.00},
		{"Division by zero", 8, 0, 0.00},
	} // Test cases
	for _, test := range tests {
		t.Run(test.test, func(t *testing.T) {
			val, err := tool.Execute(ctx, fmt.Sprintf(`{"a": %f, "b": %f}`, test.a, test.b))
			assert.Nil(t, err)
			resultVal, _ := strconv.ParseFloat(val, 64)
			assert.Equal(t, fmt.Sprintf("%.2f", test.result), fmt.Sprintf("%.2f", resultVal))
			assert.Equal(t, test.result, calcTools.Divide(ctx, test.a, test.b))
		})
	}
}

func TestCalculatorTools_Tools_Modulus(t *testing.T) {
	calcTools := &CalculatorTools{
		EnableModulus: true,
	}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	assert.NotNil(t, tools)

	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		test   string
		a, b   int
		result int
	}{
		{"Modulus of positive numbers", 5, 2, 1},
		{"Modulus of negative numbers", -5, -2, -1},
		{"Modulus with zero", 8, 0, 0},
		{"Modulus of zero", 0, 3, 0},
	} // Test cases
	for _, test := range tests {
		t.Run(test.test, func(t *testing.T) {
			val, err := tool.Execute(ctx, fmt.Sprintf(`{"a": %d, "b": %d}`, test.a, test.b))
			assert.Nil(t, err)
			resultVal, _ := strconv.Atoi(val)
			assert.Equal(t, test.result, resultVal)
			assert.Equal(t, test.result, calcTools.Modulus(ctx, test.a, test.b))
		})
	}
}

func TestCalculatorTools_Tools_Exponentiate(t *testing.T) {
	calcTools := &CalculatorTools{EnableExponentiate: true}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		name                string
		base, exp, expected float64
	}{
		{"2^3", 2, 3, 8.0},
		{"3^2", 3, 2, 9.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf(`{"base": %f, "exp": %f}`, tt.base, tt.exp)
			val, err := tool.Execute(ctx, query)
			assert.Nil(t, err)
			res, _ := strconv.ParseFloat(val, 64)
			assert.Equal(t, tt.expected, calcTools.Exponentiate(ctx, tt.base, tt.exp))
			assert.Equal(t, fmt.Sprintf("%.2f", tt.expected), fmt.Sprintf("%.2f", res))
		})
	}
}

func TestCalculatorTools_Tools_Factorial(t *testing.T) {
	calcTools := &CalculatorTools{EnableFactorial: true}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		name        string
		n, expected int
	}{
		{"Factorial of 5", 5, 120},
		{"Factorial of 0", 0, 1},
		{"Factorial of negative", -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf(`{"n": %d}`, tt.n)
			val, err := tool.Execute(ctx, query)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, calcTools.Factorial(ctx, tt.n))
			res, _ := strconv.Atoi(val)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestCalculatorTools_Tools_IsPrime(t *testing.T) {
	calcTools := &CalculatorTools{EnableIsPrime: true}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		name        string
		n           int
		expected    bool
		expectedStr string
	}{
		{"Prime number", 7, true, "true"},
		{"Non-prime positive", 8, false, "false"},
		{"Negative number", -1, false, "false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf(`{"n": %d}`, tt.n)
			val, err := tool.Execute(ctx, query)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, calcTools.IsPrime(ctx, tt.n))
			assert.Equal(t, tt.expectedStr, val)
		})
	}
}

func TestCalculatorTools_Tools_SquareRoot(t *testing.T) {
	calcTools := &CalculatorTools{EnableSquareRoot: true}
	tools := calcTools.Tools()
	assert.Equal(t, 1, len(tools))
	tool := tools[0]
	ctx := context.Background()

	tests := []struct {
		name        string
		x, expected float64
	}{
		{"Square root of positive", 9, 3.0},
		{"Square root of zero", 0, 0.0},
		{"Square root of negative", -4, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf(`{"x": %f}`, tt.x)
			val, err := tool.Execute(ctx, query)
			assert.Nil(t, err)
			res, _ := strconv.ParseFloat(val, 64)
			assert.Equal(t, tt.expected, calcTools.SquareRoot(ctx, tt.x))
			assert.Equal(t, fmt.Sprintf("%.2f", tt.expected), fmt.Sprintf("%.2f", res))
		})
	}
}
