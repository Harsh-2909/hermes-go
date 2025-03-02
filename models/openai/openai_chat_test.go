package models

import (
	"testing"
)

func TestOpenAIChatInit(t *testing.T) {
	// Test missing ApiKey
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing ApiKey")
		}
	}()
	model := OpenAIChat{Id: "test-model"}
	model.Init()

	// Test missing Id
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing Id")
		}
	}()
	model = OpenAIChat{ApiKey: "test-key"}
	model.Init()

	// Test defaults
	model = OpenAIChat{ApiKey: "test-key", Id: "test-model"}
	model.Init()
	if model.Temperature != 0.5 {
		t.Errorf("Expected default Temperature 0.5, got %f", model.Temperature)
	}
}

// Note: For ChatCompletion and ChatCompletionStream, you'd typically mock the HTTP client.
// Here's a simple example; expand with httptest.Server for full API mocking if needed.
func TestChatCompletion(t *testing.T) {
	// This requires mocking the OpenAI client, which is complex due to go-openai's internals.
	// For simplicity, test Init and basic setup here; use integration tests for full API behavior.
	model := OpenAIChat{ApiKey: "test-key", Id: "test-model"}
	model.Init()
	// Further testing requires mocking the HTTP response, omitted for brevity.
}
