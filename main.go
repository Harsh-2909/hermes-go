package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/Harsh-2909/hermes-go/models"
	openai "github.com/Harsh-2909/hermes-go/models/openai"
	"github.com/Harsh-2909/hermes-go/tools"

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

	// Building tool manually
	// params := map[string]interface{}{
	// 	"type": "object",
	// 	"properties": map[string]interface{}{
	// 		"a": map[string]interface{}{"type": "number"},
	// 		"b": map[string]interface{}{"type": "number"},
	// 	},
	// 	"required": []string{"a", "b"},
	// }
	// tool := tools.NewTool("Calculate", "Calculate the sum of two numbers", params, func(ctx context.Context, args string) (string, error) {
	// 	var input struct{ A, B int }
	// 	if err := json.Unmarshal([]byte(args), &input); err != nil {
	// 		return "", err
	// 	}
	// 	return fmt.Sprintf("%d", input.A+input.B), nil
	// })
	// tool.Tools()[0].Execute(context.Background(), `{"a": 2, "b": 3}`)

	calcTools := &tools.CalculatorTools{
		EnableAll: true,
	}
	agent := &agent.Agent{
		Model:       model,
		Description: "You are a helpful assistant.",
		Markdown:    true,
		DebugMode:   false,
		// Tools:       []tools.ToolKit{tool},
		Tools: []tools.ToolKit{calcTools},
	}

	// Non Streaming Example
	ctx := context.Background()
	response, err := agent.Run(ctx, "Can you say hello and add 267383 and 123456?")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant:", response.Data)
	fmt.Printf("\nMessages:\n\n%+v\n", agent.Messages)

	// Streaming Example
	ctx1 := context.Background()
	stream, err := agent.RunStream(ctx1, "Can you say hello and add 267383 and 123456?")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant is thinking...")
	for resp := range stream {
		if resp.Event == "chunk" {
			fmt.Print(resp.Data)
		} else if resp.Event == "end" {
			fmt.Print("\nAssistant is Done.")
			break
		} else if resp.Event == "error" {
			fmt.Println("\nError:", resp.Data)
			break
		}
	}
	fmt.Printf("\nMessages:\n\n%+v\n", agent.Messages)

	// Multimodal Example
	// Image content
	ctx2 := context.Background()
	image := &models.Image{
		URL: "https://store.storeimages.cdn-apple.com/4668/as-images.apple.com/is/ipad-10th-gen-finish-select-202212-pink-wifi_FMT_WHH?wid=1200&hei=630&fmt=jpeg&qlt=95&.v=1670856074755",
	}
	audio := &models.Audio{
		URL: "https://file-examples-com.github.io/uploads/2017/11/file_example_MP3_700KB.mp3",
	}
	response, err = agent.Run(ctx2, "What is in this image?", image, audio)
	// response, err = agent.Run(ctx2, "What is in this image?", image)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Assistant:", response.Data)
}
