package main

import (
	"context"
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
	calcTools := &tools.CalculatorTools{
		EnableAll: true,
	}
	agent := agent.Agent{
		Model: &models.OpenAIChat{
			ApiKey: os.Getenv("OPENAI_API_KEY"),
			Id:     "gpt-4o-mini",
		},
		Markdown:  true,
		DebugMode: true,
		Tools:     []tools.ToolKit{calcTools},
	}

	// Calculate the expression
	err = agent.PrintResponse(context.Background(), "What is 10*5 then to the power of 2, do it step by step.", true, true)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}
}
