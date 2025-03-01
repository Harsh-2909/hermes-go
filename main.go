package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/Harsh-2909/hermes-go/models"

	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the OpenAI model
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := &models.OpenAIModel{
		ApiKey: apiKey,
		Id:     "gpt-4o-mini",
	}
	model.Init()

	agent := agent.Agent{
		Model:         model,
		SystemMessage: "You are a helpful assistant.",
	}
	agent.Init()

	// Send a user message and get a response
	ctx := context.Background()
	response, err := agent.RespondTo(ctx, "Hello, how are you?")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant:", response)
}
