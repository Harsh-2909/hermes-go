package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	models "github.com/Harsh-2909/hermes-go/models/openai"

	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the OpenAI model
	model := &models.OpenAIChat{
		ApiKey: os.Getenv("OPENAI_API_KEY"),
		Id:     "gpt-4o-mini",
	}
	model.Init()

	agent := &agent.Agent{
		Model:         model,
		SystemMessage: "You are a helpful assistant.",
	}
	agent.Init()

	// Non Streaming Example
	// ctx := context.Background()
	// response, err := agent.Run(ctx, "Hello, how are you?")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// fmt.Println("Assistant:", response.Data)

	// Streaming Example
	ctx1 := context.Background()
	stream, err := agent.RunStream(ctx1, "What is the meaning of life?")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant is thinking...")
	for resp := range stream {
		if resp.Event == "chunk" {
			fmt.Print(resp.Data)
		} else if resp.Event == "complete" {
			fmt.Print("\nAssistant is Done.")
			break
		} else if resp.Event == "error" {
			fmt.Println("\nError:", resp.Data)
			break
		}
	}
}
