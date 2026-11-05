package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type ReadFileTool struct {
	workDir string
}

// Name implements [BaseTool].
func (r *ReadFileTool) Name() string {
	return "read_file"
}

// Definition implements [BaseTool].
func (r *ReadFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        r.Name(),
		Description: "Read the content of a file.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The path of the file to read.",
				},
			},
			"required": []string{"path"},
		},
	}
}

type readFileArgs struct {
	Path string `json:"path"`
}

// Execute implements [BaseTool].
func (r *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	// 1. Parse the input arguments.
	var input readFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("Parameter parsing failed: %w", err)
	}

	// 2. Construct the full path to the file.
	fullPath := filepath.Join(r.workDir, input.Path)

	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	// 3. Read the content of the file.
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %w", err)
	}

	const maxContentLength = 4096
	if len(content) > maxContentLength {
		truncatedContent := content[:maxContentLength]
		fmt.Printf("%s \n\n File content is too long. Returning the first %d characters.", truncatedContent, maxContentLength)
		return string(truncatedContent), nil
	}
	return string(content), nil
}

func NewReadFileTool(workDir string) *ReadFileTool {
	return &ReadFileTool{
		workDir: workDir,
	}
}
