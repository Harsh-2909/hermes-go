package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/Harsh-2909/hermes-go/models"
	openai "github.com/Harsh-2909/hermes-go/models/openai"

	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize the OpenAI model
	model := &openai.OpenAIChat{
		ApiKey: os.Getenv("OPENAI_API_KEY"),
		Id:     "gpt-4o-mini",
	}
	model.Init()

	agent := &agent.Agent{
		Model:       model,
		Description: "You are a helpful assistant.",
		Markdown:    true,
	}
	agent.Init()

	// Non Streaming Example
	ctx := context.Background()
	response, err := agent.Run(ctx, "Hello, how are you?")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant:", response.Data)

	// Streaming Example
	// ctx1 := context.Background()
	// stream, err := agent.RunStream(ctx1, "What is the meaning of life?")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// fmt.Println("Assistant is thinking...")
	// for resp := range stream {
	// 	if resp.Event == "chunk" {
	// 		fmt.Print(resp.Data)
	// 	} else if resp.Event == "complete" {
	// 		fmt.Print("\nAssistant is Done.")
	// 		break
	// 	} else if resp.Event == "error" {
	// 		fmt.Println("\nError:", resp.Data)
	// 		break
	// 	}
	// }

	// Multimodal Example
	// Image content
	ctx2 := context.Background()
	image := &models.Image{
		URL: "https://store.storeimages.cdn-apple.com/4668/as-images.apple.com/is/ipad-10th-gen-finish-select-202212-pink-wifi_FMT_WHH?wid=1200&hei=630&fmt=jpeg&qlt=95&.v=1670856074755",
	}
	response, err = agent.Run(ctx2, "What is in this image?", image)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant:", response.Data)
}
