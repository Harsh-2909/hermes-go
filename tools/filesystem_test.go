package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemTools_WriteFile_Defaults(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	ftools := &FileSystemTools{
		EnableWriteFile:  true,
		TargetDirectory:  tempDir,
		DefaultExtension: "txt",
	}
	content := "Hello, world!"
	msg, err := ftools.WriteFile(ctx, content, "", "", "")
	assert.NoError(t, err)
	assert.True(t, strings.Contains(msg, "Successfully wrote file to:"))
	filePath := strings.TrimPrefix(msg, "Successfully wrote file to: ")
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
	os.Remove(filePath)
}

func TestFileSystemTools_WriteFile_WithFilename(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	ftools := &FileSystemTools{
		EnableWriteFile:  true,
		TargetDirectory:  tempDir,
		DefaultExtension: "txt",
	}
	content := "Explicit filename test"
	filename := "testfile.md" // provided filename with extension
	msg, err := ftools.WriteFile(ctx, content, filename, "", "")
	assert.NoError(t, err)
	expectedPath := filepath.Join(tempDir, "testfile.md")
	assert.True(t, strings.Contains(msg, expectedPath))
	data, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
	os.Remove(expectedPath)
}

func TestFileSystemTools_ReadFile_Exists(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	ftools := &FileSystemTools{
		EnableWriteFile:  true,
		EnableReadFile:   true,
		TargetDirectory:  tempDir,
		DefaultExtension: "txt",
	}
	content := "Read test content"
	filename := "readtest"
	_, err := ftools.WriteFile(ctx, content, filename, "", "")
	assert.NoError(t, err)
	expectedPath := filepath.Join(tempDir, filename+".txt")
	readContent, err := ftools.ReadFile(ctx, filename+".txt", "")
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)
	os.Remove(expectedPath)
}

func TestFileSystemTools_ReadFile_NotExist(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	ftools := &FileSystemTools{
		EnableReadFile:  true,
		TargetDirectory: tempDir,
	}
	filename := "nonexistentfile.txt"
	msg, err := ftools.ReadFile(ctx, filename, "")
	assert.NoError(t, err)
	assert.True(t, strings.Contains(msg, "File not found:"))
}
