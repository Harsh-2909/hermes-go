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
	fileTools := &tools.FileSystemTools{
		EnableAll:        true,
		TargetDirectory:  "./files",
		DefaultExtension: "txt",
	}
	agent := agent.Agent{
		Model: &models.OpenAIChat{
			ApiKey: os.Getenv("OPENAI_API_KEY"),
			Id:     "gpt-4o-mini",
		},
		Description: "You are an assistant with access to file system tools.",
		Markdown:    true,
		DebugMode:   true,
		Tools:       []tools.ToolKit{fileTools},
	}

	// Create a file
	err = agent.PrintResponse(context.Background(), "Create a file with the content 'Hello, World!' in the current directory with file name hello.txt.", true, true)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Read the file
	err = agent.PrintResponse(context.Background(), "Read the file hello.txt and display its content.", true, true)
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}
}
