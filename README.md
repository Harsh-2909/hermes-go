# Hermes-Go

Hermes-Go is a powerful Go-based AI Agent framework inspired by LangChain and Agno. It provides a flexible and intuitive way to interact with language models, build AI-powered applications, and create custom tools for enhanced AI capabilities.

## Features

- 🤖 Easy-to-use AI Agent implementation
- 🔧 Custom tool creation and integration
- 📡 Support for streaming responses
- 🖼️ Multimodal capabilities (images, audio)
- 🔌 Modular design for easy extensibility
- ✨ Markdown formatting support
- 🛠️ Built-in debugging mode

## Requirements

- Go 1.23 or higher
- OpenAI API key

## Installation

```bash
go get github.com/Harsh-2909/hermes-go
```

## Environment Setup

Create a `.env` file in your project root and add your OpenAI API key:

```env
OPENAI_API_KEY=your_api_key_here
```

## Supported Models

Currently, Hermes-Go supports the following models:
- OpenAI chat completion models

More models will be added in the future.

## Future Plans

- Make tool creation more user-friendly
- Support for more AI models
- Enhanced multimodal capabilities
- Tools package with common tools already implemented
- More examples and documentation
- Database integration for persistent data storage (Memory Layer)
- Built-in support for Knowledge structs for easy data retrieval

## Usage Examples

Examples for running Hermes in present in ./examples folder.

### Basic Setup

```go
import (
    "github.com/Harsh-2909/hermes-go/agent"
    "github.com/Harsh-2909/hermes-go/models/openai"
    "github.com/Harsh-2909/hermes-go/tools"
)

// Initialize the OpenAI model
model := &openai.OpenAIChat{
    ApiKey: os.Getenv("OPENAI_API_KEY"),
    Id:     "gpt-4-mini",
}

// Create a new agent
agent := &agent.Agent{
    Model:       model,
    Description: "You are a helpful assistant.",
    Markdown:    true,
    Tools:       []tools.ToolKit{},
}
```

### Creating Custom Tools

```go
// Define tool parameters
params := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "a": map[string]interface{}{"type": "number"},
        "b": map[string]interface{}{"type": "number"},
    },
    "required": []string{"a", "b"},
}

// Create a calculation tool
tool := tools.NewTool("Calculate", "Calculate the sum of two numbers", params, func(ctx context.Context, args string) (string, error) {
    var input struct{ A, B int }
    if err := json.Unmarshal([]byte(args), &input); err != nil {
        return "", err
    }
    return fmt.Sprintf("%d", input.A+input.B), nil
})

// Add tool to agent
agent.Tools = []tools.ToolKit{tool}
```

### Non-Streaming Example

```go
ctx := context.Background()
response, err := agent.Run(ctx, "Can you say hello and add 267383 and 123456?")
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println("Assistant:", response.Data)
```

### Streaming Example

```go
ctx := context.Background()
stream, err := agent.RunStream(ctx, "Can you say hello and add 267383 and 123456?")
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
```

### Multimodal Example

```go
// Image processing example
image := &models.Image{
    URL: "https://example.com/image.jpg",
}

// Process image with agent
response, err = agent.Run(ctx, "What is in this image?", image)
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Println("Assistant:", response.Data)
```

## Debug Mode

Enable debug mode to get detailed information about the agent's operations:

```go
agent := &agent.Agent{
    // ... other configurations ...
    DebugMode: true,
}
```