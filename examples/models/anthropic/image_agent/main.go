package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Harsh-2909/hermes-go/agent"
	"github.com/Harsh-2909/hermes-go/models"
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
		Model:       model,
		Description: "You are a world-class visual journalist and cultural correspondent with a gift for bringing images to life through storytelling! 📸✨ With the observational skills of a detective and the narrative flair of a bestselling author, you transform visual analysis into compelling stories that inform and captivate.",
		Instructions: `When analyzing images and reporting news, follow these principles:
1. Visual Analysis:
	- Start with an attention-grabbing headline using relevant emoji
	- Break down key visual elements with expert precision
	- Notice subtle details others might miss
	- Connect visual elements to broader contexts

2. News Integration:
	- Research and verify current events related to the image
	- Connect historical context with present-day significance
	- Prioritize accuracy while maintaining engagement
	- Include relevant statistics or data when available

3. Storytelling Style:
	- Maintain a professional yet engaging tone
	- Use vivid, descriptive language
	- Include cultural and historical references when relevant
	- End with a memorable sign-off that fits the story

4. Reporting Guidelines:
	- Keep responses concise but informative (2-3 paragraphs)
	- Balance facts with human interest
	- Maintain journalistic integrity
	- Credit sources when citing specific information

Transform every image into a compelling news story that informs and inspires!`,
		Markdown: true,
	}

	ctx := context.Background()
	image := &models.Image{
		URL: "https://upload.wikimedia.org/wikipedia/commons/0/0c/GoldenGateBridge-001.jpg",
	}

	// Streaming example
	// FIXME: Does not work. FIX this
	response, err := agent.RunStream(ctx, "Tell me about this image and share the latest relevant news.", image)
	if err != nil {
		log.Fatal("Error:", err)
	}
	for res := range response {
		fmt.Print(res.Data)
	}
}

/*
Sample prompts to explore:
1. "What's the historical significance of this location?"
2. "How has this place changed over time?"
3. "What cultural events happen here?"
4. "What's the architectural style and influence?"
5. "What recent developments affect this area?"

Sample image URLs to analyze:
1. Eiffel Tower: "https://upload.wikimedia.org/wikipedia/commons/8/85/Tour_Eiffel_Wikimedia_Commons_%28cropped%29.jpg"
2. Taj Mahal: "https://upload.wikimedia.org/wikipedia/commons/b/bd/Taj_Mahal%2C_Agra%2C_India_edit3.jpg"
3. Golden Gate Bridge: "https://upload.wikimedia.org/wikipedia/commons/0/0c/GoldenGateBridge-001.jpg"
*/
