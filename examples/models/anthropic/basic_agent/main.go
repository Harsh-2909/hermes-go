package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	anthropic "github.com/Harsh-2909/hermes-go/models/anthropic"

	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the model
	model := &anthropic.Claude{
		ApiKey:      os.Getenv("ANTHROPIC_API_KEY"),
		Id:          "claude-3-5-haiku-latest",
		Temperature: 1.0,
	}

	// Create a new agent
	agent := &agent.Agent{
		Model: model,
		Instructions: []string{
			"You are an enthusiastic news reporter with a flair for storytelling! 🗽",
			"Think of yourself as a mix between a witty comedian and a sharp journalist.",
			"Your style guide:",
			"- Start with an attention-grabbing headline using emoji",
			"- Share news with enthusiasm and NYC attitude",
			"- Keep your responses concise but entertaining",
			"- Throw in local references and NYC slang when appropriate",
			"- End with a catchy sign-off like 'Back to you in the studio!' or 'Reporting live from the Big Apple!'",
			"Remember to verify all facts while keeping that NYC energy high!",
		},
		Markdown: true,
	}

	// Non-streaming example
	ctx := context.Background()
	response, err := agent.Run(ctx, "What's the latest scoop in NYC?")
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println("Assistant:")
	fmt.Println(response.Data)
}
