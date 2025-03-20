package agent

import (
	"context"
	"testing"
	"time"

	"github.com/Harsh-2909/hermes-go/models"
	"github.com/Harsh-2909/hermes-go/tools"

	"github.com/stretchr/testify/assert"
)

// MockModel is a mock implementation of the Model interface for testing.
type MockModel struct {
	tools []tools.Tool
}

func (m *MockModel) Init() {}
func (m *MockModel) SetTools(tools []tools.Tool) {
	m.tools = tools
}
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

// MockTool is a simple implementation of the Tool interface for testing
func createMockTool(name string) tools.Tool {
	return tools.Tool{
		Name:        name,
		Description: "Test tool " + name,
		Parameters:  map[string]interface{}{"param": "test"},
		Execute: func(ctx context.Context, args string) (string, error) {
			return "Executed " + name, nil
		},
	}
}

// MockToolKit is a simple implementation of the ToolKit interface for testing
type MockToolKit struct {
	ToolNames []string
}

func (tk *MockToolKit) Tools() []tools.Tool {
	var toolList []tools.Tool
	for _, name := range tk.ToolNames {
		toolList = append(toolList, createMockTool(name))
	}
	return toolList
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
	assert.Len(t, agent.Messages, 1, "Expected 1 system message in Messages after Init")
	assert.Equal(t, "system", agent.Messages[0].Role, "Expected message of role `system` in Messages after Init")
}

func TestGetSystemMessage(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		goal         string
		role         string
		markdown     bool
		instructions interface{}
		expected     string
	}{
		{
			name:         "All fields with markdown",
			description:  "Test agent",
			goal:         "Test goal",
			role:         "Test role",
			markdown:     true,
			instructions: nil,
			expected:     "Test agent\n\n<your_goal>\nTest goal\n</your_goal>\n\n<your_role>\nTest role\n</your_role>\n\n<additional_information>\n- Use markdown to format your answers.\n</additional_information>\n\n",
		},
		{
			name:         "All fields without markdown",
			description:  "Test agent",
			goal:         "Test goal",
			role:         "Test role",
			markdown:     false,
			instructions: nil,
			expected:     "Test agent\n\n<your_goal>\nTest goal\n</your_goal>\n\n<your_role>\nTest role\n</your_role>\n\n",
		},
		{
			name:         "With string instructions",
			description:  "Test agent",
			goal:         "Test goal",
			role:         "Test role",
			markdown:     false,
			instructions: "Remember to be concise.",
			expected:     "Test agent\n\n<your_goal>\nTest goal\n</your_goal>\n\n<your_role>\nTest role\n</your_role>\n\n<instructions>\nRemember to be concise.\n</instructions>\n\n",
		},
		{
			name:         "With slice instructions and markdown",
			description:  "Agent",
			goal:         "Goal A",
			role:         "Role A",
			markdown:     true,
			instructions: []string{"Step 1", "Step 2"},
			expected:     "Agent\n\n<your_goal>\nGoal A\n</your_goal>\n\n<your_role>\nRole A\n</your_role>\n\n<instructions>\n- Step 1\n- Step 2\n</instructions>\n\n<additional_information>\n- Use markdown to format your answers.\n</additional_information>\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := Agent{
				Model:        &MockModel{},
				Description:  tc.description,
				Goal:         tc.goal,
				Role:         tc.role,
				Markdown:     tc.markdown,
				Instructions: tc.instructions,
			}
			agent.Init()
			msg := agent.getSystemMessage()
			assert.Equal(t, "system", msg.Role, "Expected message of role `system`")
			assert.Equal(t, tc.expected, msg.Content, "Expected predefined system message")
		})
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

func TestProcessTools(t *testing.T) {
	tests := []struct {
		name     string
		tools    []tools.ToolKit
		expected int
	}{
		{
			name:     "Empty tool list",
			tools:    []tools.ToolKit{},
			expected: 0,
		},
		{
			name:     "Single tool",
			tools:    []tools.ToolKit{createMockTool("tool1")},
			expected: 1,
		},
		{
			name:     "Multiple tools as Tool implementations",
			tools:    []tools.ToolKit{createMockTool("tool1"), createMockTool("tool2")},
			expected: 2,
		},
		{
			name: "Mixed Tool and ToolKit implementations",
			tools: []tools.ToolKit{
				createMockTool("tool1"),
				&MockToolKit{ToolNames: []string{"kit-tool1", "kit-tool2"}},
			},
			expected: 3,
		},
		{
			name: "Multiple ToolKit implementations",
			tools: []tools.ToolKit{
				&MockToolKit{ToolNames: []string{"kit1-tool1", "kit1-tool2"}},
				&MockToolKit{ToolNames: []string{"kit2-tool1", "kit2-tool2", "kit2-tool3"}},
			},
			expected: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := Agent{Model: &MockModel{}, Tools: tc.tools}
			processedTools := agent.processTools()

			assert.Equal(t, tc.expected, len(processedTools), "Processed tools count should match expected")

			// Check that all tools were properly processed with correct names
			if tc.expected > 0 {
				toolNames := make(map[string]bool)
				for _, tool := range processedTools {
					toolNames[tool.Name] = true
				}

				// Verify all expected tools are present
				for _, toolkit := range tc.tools {
					if tool, ok := toolkit.(tools.Tool); ok {
						assert.True(t, toolNames[tool.Name], "Tool %s should be in processed tools", tool.Name)
					} else {
						for _, tool := range toolkit.Tools() {
							assert.True(t, toolNames[tool.Name], "Tool %s should be in processed tools", tool.Name)
						}
					}
				}
			}
		})
	}
}

func TestGetAllTools(t *testing.T) {
	// Test case 1: Empty tools list
	t.Run("Empty tools list", func(t *testing.T) {
		agent := Agent{Tools: []tools.ToolKit{}}
		tools := agent.GetAllTools()
		assert.Empty(t, tools, "Tools should be empty")
	})

	// Test case 2: First call with no cached tools
	t.Run("First call with no cached tools", func(t *testing.T) {
		agent := Agent{
			Tools: []tools.ToolKit{
				createMockTool("tool1"),
				createMockTool("tool2"),
			},
		}

		// Verify _tools is empty before call
		assert.Empty(t, agent._tools, "_tools should be empty before first call")

		// Get all tools should process and cache
		tools := agent.GetAllTools()
		assert.Len(t, tools, 2, "Should return 2 tools")
		assert.Len(t, agent._tools, 2, "_tools should be cached with 2 tools")

		// Verify tool names
		names := []string{tools[0].Name, tools[1].Name}
		assert.Contains(t, names, "tool1")
		assert.Contains(t, names, "tool2")
	})

	// Test case 3: Second call should use cached tools
	t.Run("Second call should use cached tools", func(t *testing.T) {
		// Setup agent with tools
		agent := Agent{
			Tools: []tools.ToolKit{
				createMockTool("tool1"),
				createMockTool("tool2"),
			},
		}

		// First call to cache tools
		firstTools := agent.GetAllTools()
		assert.Len(t, firstTools, 2)

		// Modify Tools but keep _tools cache
		agent.Tools = append(agent.Tools, createMockTool("tool3"))

		// Second call should use cached version
		secondTools := agent.GetAllTools()
		assert.Len(t, secondTools, 2, "Should still return 2 tools from cache")
		assert.Equal(t, firstTools, secondTools, "Should return same tools from cache")

		// Clear cache and call again
		agent._tools = nil
		thirdTools := agent.GetAllTools()
		assert.Len(t, thirdTools, 3, "Should return 3 tools after cache cleared")
	})
}

func TestAddToolToModel(t *testing.T) {
	tests := []struct {
		name     string
		tools    []tools.ToolKit
		expected int
	}{
		{
			name:     "Empty tool list",
			tools:    []tools.ToolKit{},
			expected: 0,
		},
		{
			name:     "Single tool",
			tools:    []tools.ToolKit{createMockTool("tool1")},
			expected: 1,
		},
		{
			name: "Mixed Tool and ToolKit implementations",
			tools: []tools.ToolKit{
				createMockTool("tool1"),
				&MockToolKit{ToolNames: []string{"kit-tool1", "kit-tool2"}},
			},
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockModel := &MockModel{}
			agent := Agent{
				Model: mockModel,
				Tools: tc.tools,
			}

			// Call the method being tested
			agent.addToolToModel()

			// Verify the correct tools were set on the model
			if tc.expected == 0 {
				assert.Empty(t, mockModel.tools, "No tools should be set for empty tool list")
			} else {
				assert.Len(t, mockModel.tools, tc.expected, "Model should receive correct number of tools")

				// Verify all tools were set with the correct names
				allTools := agent.GetAllTools()
				assert.Equal(t, allTools, mockModel.tools, "All tools should be passed to the model")
			}
		})
	}
}
