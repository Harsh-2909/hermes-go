package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	models "github.com/Harsh-2909/hermes-go/models/openai"
	"github.com/Harsh-2909/hermes-go/tools"

	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Building tool manually
	// Writing a calculator add tool
	params := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"a": map[string]interface{}{"type": "number"},
			"b": map[string]interface{}{"type": "number"},
		},
		"required": []string{"a", "b"},
	}
	tool := tools.NewTool("Calculate", "Calculate the sum of two numbers", params, func(ctx context.Context, args string) (string, error) {
		var input struct{ A, B int }
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", input.A+input.B), nil
	})

	// Defining Agent
	agent := agent.Agent{
		Model: &models.OpenAIChat{
			ApiKey: os.Getenv("OPENAI_API_KEY"),
			Id:     "gpt-4o-mini",
		},
		Description: "You are a computing assistant which helps users with their math queries.",
		Markdown:    true,
		Tools:       []tools.ToolKit{tool},
	}

	err = agent.PrintResponse(context.Background(), "What is the sum of 123456 and 654321", true, true)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}
}
