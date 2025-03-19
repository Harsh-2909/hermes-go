package agent

import (
	"context"
	"testing"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/Harsh-2909/hermes-go/tools"
)

// MockModel is a mock implementation of the Model interface for testing.
type MockModel struct{}

func (m *MockModel) Init()                       {}
func (m *MockModel) SetTools(tools []tools.Tool) {}
func (m *MockModel) ChatCompletion(ctx context.Context, messages []models.Message) (models.ModelResponse, error) {
	return models.ModelResponse{
		Event:     "complete",
		Data:      "Mock response",
		CreatedAt: time.Now(),
	}, nil
}
func (m *MockModel) ChatCompletionStream(ctx context.Context, messages []models.Message) (chan models.ModelResponse, error) {
	ch := make(chan models.ModelResponse)
	go func() {
		defer close(ch)
		ch <- models.ModelResponse{Event: "chunk", Data: "Mock chunk", CreatedAt: time.Now()}
		ch <- models.ModelResponse{Event: "end", CreatedAt: time.Now()}
	}()
	return ch, nil
}

func TestAgentInit(t *testing.T) {
	// Test panic on nil Model
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when Model is nil")
		}
	}()
	agent := Agent{}
	agent.Init()

	// Test successful initialization
	agent = Agent{Model: &MockModel{}, Description: "Test agent"}
	agent.Init()
	if len(agent.Messages) != 1 || agent.Messages[0].Role != "system" {
		t.Errorf("Expected system message in Messages after Init")
	}
}

func TestGetSystemMessage(t *testing.T) {
	agent := Agent{
		Model:       &MockModel{},
		Description: "Test agent",
		Goal:        "Test goal",
		Role:        "Test role",
		Markdown:    true,
	}
	agent.Init()
	msg := agent.getSystemMessage()
	expected := "Test agent\n\n<your_goal>\nTest goal\n</your_goal>\n\n<your_role>\nTest role\n</your_role>\n\n<additional_information>\n- Use markdown to format your answers.\n</additional_information>\n\n"
	if msg.Role != "system" || msg.Content != expected {
		t.Errorf("Expected system message '%s', got '%s'", expected, msg.Content)
	}
}

func TestAddMessage(t *testing.T) {
	// Text only
	agent := Agent{Model: &MockModel{}}
	agent.Init()
	agent.AddMessage("user", "Hello", nil)
	if len(agent.Messages) != 1 || agent.Messages[0].Role != "user" || agent.Messages[0].Content != "Hello" || len(agent.Messages[0].Images) != 0 || len(agent.Messages[0].Audios) != 0 {
		t.Errorf("Expected message {user, Hello, nil}, got %+v", agent.Messages)
	}

	// Text with image
	image := &models.Image{URL: "http://example.com/image.png"}
	agent.AddMessage("user", "Hello with image", []models.Media{image})
	if len(agent.Messages) != 2 || agent.Messages[1].Role != "user" || agent.Messages[1].Content != "Hello with image" || len(agent.Messages[1].Images) != 1 {
		t.Errorf("Expected message {user, Hello with image, image}, got %+v", agent.Messages)
	}

	// Text with audio
	audio := &models.Audio{URL: "http://example.com/audio.mp3"}
	agent.AddMessage("user", "Hello with audio", []models.Media{audio})
	if len(agent.Messages) != 3 || agent.Messages[2].Role != "user" || agent.Messages[2].Content != "Hello with audio" || len(agent.Messages[2].Audios) != 1 {
		t.Errorf("Expected message {user, Hello with audio, audio}, got %+v", agent.Messages)
	}

	// Text with image and audio
	agent.AddMessage("user", "Hello with image and audio", []models.Media{image, audio})
	if len(agent.Messages) != 4 || agent.Messages[3].Role != "user" || agent.Messages[3].Content != "Hello with image and audio" || len(agent.Messages[3].Images) != 1 || len(agent.Messages[3].Audios) != 1 {
		t.Errorf("Expected message {user, Hello with image and audio, image, audio}, got %+v", agent.Messages)
	}
}

func TestRun(t *testing.T) {
	agent := Agent{Model: &MockModel{}}
	agent.Init()
	resp, err := agent.Run(context.Background(), "Hi there")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.Data != "Mock response" || len(agent.Messages) != 2 { // System + Assistant
		t.Errorf("Expected response 'Mock response' and 2 messages, got %+v, %d messages", resp, len(agent.Messages))
		t.Errorf("Messages: %+v", agent.Messages)
	}
}

func TestRunStream(t *testing.T) {
	agent := Agent{Model: &MockModel{}}
	agent.Init()
	ch, err := agent.RunStream(context.Background(), "Stream me")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	count := 0
	for resp := range ch {
		if resp.Event == "chunk" && resp.Data != "Mock chunk" {
			t.Errorf("Expected chunk 'Mock chunk', got '%s'", resp.Data)
		}
		count++
	}
	if count != 2 { // chunk + end
		t.Errorf("Expected 2 events, got %d", count)
	}
}
