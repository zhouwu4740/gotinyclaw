package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type WriteFileTool struct {
	workDir string
}

// Name implements [BaseTool].
func (w *WriteFileTool) Name() string {
	return "write_file"
}

// Definition implements [BaseTool].
func (w *WriteFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        w.Name(),
		Description: "Write the content to a file.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The path of the file to write.",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The content to write to the file.",
				},
			},
			"required": []string{"path", "content"},
		},
	}
}

type writeFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Execute implements [BaseTool].
func (w *WriteFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	// 1. Parse the input arguments.
	var input writeFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// 2. Construct the full path to the file.
	fullPath := filepath.Join(w.workDir, input.Path)

	// 3. Write the content to the file.
	// 3.1 Create directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	// 3.2 Write the content to the file.
	if err := os.WriteFile(fullPath, []byte(input.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return fmt.Sprintf("Wrote file to %s\n", fullPath), nil
}

func NewWriteFileTool(workDir string) *WriteFileTool {
	return &WriteFileTool{
		workDir: workDir,
	}
}
