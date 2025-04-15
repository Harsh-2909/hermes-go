package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Harsh-2909/hermes-go/utils"
	"github.com/google/uuid"
)

// FileSystemTools provides tools for interacting with the local file system.
type FileSystemTools struct {
	EnableWriteFile  bool   // Enable the write_file tool
	EnableReadFile   bool   // Enable the read_file tool
	EnableAll        bool   // Enable all tools if true
	TargetDirectory  string // Default directory for file operations
	DefaultExtension string // Default file extension (e.g., "txt")
}

// Tools returns a list of available tools based on enable flags.
func (f *FileSystemTools) Tools() []Tool {
	var tools []Tool

	if f.EnableWriteFile || f.EnableAll {
		if writeTool, err := CreateToolFromMethod(f, "WriteFile"); err == nil {
			tools = append(tools, writeTool)
		} else {
			utils.Logger.Error("Failed to create tool", "tool", "WriteFile", "error", err)
		}
	}

	if f.EnableReadFile || f.EnableAll {
		if readTool, err := CreateToolFromMethod(f, "ReadFile"); err == nil {
			tools = append(tools, readTool)
		} else {
			utils.Logger.Error("Failed to create tool", "tool", "ReadFile", "error", err)
		}
	}

	return tools
}

// WriteFile writes content to a local file.
// @param content: Content to write to the file
// @param [optional] filename: Name of the file. Defaults to UUID if not provided
// @param [optional] directory: Directory to write file to. Uses TargetDirectory if not provided
// @param [optional] extension: File extension. Uses DefaultExtension if not provided
// @return Path to the created file or error message
func (f *FileSystemTools) WriteFile(ctx context.Context, content, filename, directory, extension string) (string, error) {
	// Use defaults if parameters are empty
	if directory == "" {
		directory = f.TargetDirectory
	}
	if extension == "" {
		extension = f.DefaultExtension
	}
	if filename == "" {
		filename = uuid.New().String()
	} else {
		// Extract extension from filename if provided
		if ext := filepath.Ext(filename); ext != "" {
			extension = ext[1:] // Remove leading dot
			filename = filename[:len(filename)-len(ext)]
		}
	}

	// Ensure directory exists
	dirPath := filepath.Clean(directory)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Construct full file path
	fullFilename := fmt.Sprintf("%s.%s", filename, extension)
	filePath := filepath.Join(dirPath, fullFilename)

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return fmt.Sprintf("Successfully wrote file to: %s", filePath), nil
}

// ReadFile reads content from a local file.
// @param filename: Name of the file
// @param [optional] directory: Directory of the file. Uses TargetDirectory if not provided
// @return Content of the file or error message
func (f *FileSystemTools) ReadFile(ctx context.Context, filename, directory string) (string, error) {
	// Use default directory if not provided
	if directory == "" {
		directory = f.TargetDirectory
	}

	// Construct full file path
	filePath := filepath.Join(directory, filename)

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", filePath), nil
		}
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(data), nil
}
