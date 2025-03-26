//go:build agent_with_tool

package main

import (
	"context"
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
	calc := &tools.CalculatorTools{
		EnableAll: true,
	}
	agent := agent.Agent{
		Model: &models.OpenAIChat{
			ApiKey: os.Getenv("OPENAI_API_KEY"),
			Id:     "gpt-4o-mini",
		},
		Description: "You are a computing assistant which helps users with their math queries.",
		Markdown:    true,
		Tools:       []tools.ToolKit{calc},
		// DebugMode:   true,
	}

	fmt.Printf("Running agent...\n")
	message := "Can you find the area of a rectangle with length 23 and breadth 7?"
	fmt.Printf("User: %s\n", message)
	ch, err := agent.RunStream(context.Background(), message)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}
	fmt.Printf("Agent: ")
	for msg := range ch {
		fmt.Printf("%s", msg.Data)
	}
	fmt.Printf("\nExecution completed successfully.\n")
}
