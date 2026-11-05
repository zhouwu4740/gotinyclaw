package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type EditFileTool struct {
	workingDir string
}

// Name implements [BaseTool].
func (e *EditFileTool) Name() string {
	return "edit_file"
}

// Definition implements [BaseTool].
func (e *EditFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        e.Name(),
		Description: "Edit the content of a file, supports partial updates.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The path of the file to edit.",
				},
				"old_content": map[string]any{
					"type":        "string",
					"description": "The original content needed to be updated.",
				},
				"new_content": map[string]any{
					"type":        "string",
					"description": "The new content to update.",
				},
			},
			"required": []string{"path", "old_content", "new_content"},
		},
	}
}

type editFileArgs struct {
	Path       string `json:"path"`
	OldContent string `json:"old_content"`
	NewContent string `json:"new_content"`
}

// Execute implements [BaseTool].
func (t *EditFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	// 1. Parse the input arguments.
	var input editFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("execute edit_file tool failed, invalid input: %w", err)
	}

	fullPath := filepath.Join(t.workingDir, input.Path)

	// 2. Read the content of the file.
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("execute edit_file tool failed, failed to read file: %w", err)
	}
	content := string(contentBytes)

	// 3. Replace the old content with the new content.
	newContent, err := fuzzyReplace(content, input.OldContent, input.NewContent)
	if err != nil {
		return "", fmt.Errorf("execute edit_file tool failed, failed to replace content: %w", err)
	}

	// 4. Write the new content back to the file.
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("execute edit_file tool failed, failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully updated file: %s\n", fullPath), nil
}

func fuzzyReplace(original string, oldContent string, newContent string) (string, error) {

	// 1. 精确匹配
	count := strings.Count(original, oldContent)
	if count == 1 {
		return strings.Replace(original, oldContent, newContent, 1), nil
	}
	if count > 1 {
		return "", fmt.Errorf("old content appears %d times in the original text, please provide more specific content to determine which one to replace", count)
	}

	// 2. 转换为统一格式匹配
	normalizedOriginal := strings.ReplaceAll(original, "\r\n", "\n")
	normalizedOldContent := strings.ReplaceAll(oldContent, "\r\n", "\n")

	count = strings.Count(normalizedOriginal, normalizedOldContent)
	if count == 1 {
		return strings.Replace(normalizedOriginal, normalizedOldContent, newContent, 1), nil
	}

	// 3. 去除首尾空格匹配
	trimmedOld := strings.TrimSpace(normalizedOldContent)
	if trimmedOld != "" {
		count = strings.Count(normalizedOriginal, trimmedOld)
		if count == 1 {
			return strings.Replace(normalizedOriginal, trimmedOld, newContent, 1), nil
		}
	}

	// 4. 逐行匹配替换
	return lineByLineReplace(original, oldContent, newContent)
}

// lineByLineReplace 逐行匹配替换, 将文本按行切割, 去除首尾空格后进行滑动窗口匹配
func lineByLineReplace(original string, oldContent string, newContent string) (string, error) {
	originalLines := strings.Split(original, "\n")
	oldContentLines := strings.Split(strings.TrimSpace(oldContent), "\n")

	if len(oldContentLines) == 0 || len(originalLines) < len(oldContentLines) {
		return "", fmt.Errorf("doesn't contain the old content")
	}

	// 清理 oldContentLines 每行空格
	for i := range oldContentLines {
		oldContentLines[i] = strings.TrimSpace(oldContentLines[i])
	}

	// 滑动窗口在原始文本中寻找匹配块
	matchCount := 0
	matchStartIndex := -1
	matchEndIndex := -1

	for i := 0; i <= len(originalLines)-len(oldContentLines); i++ {
		isMatch := true
		for j := 0; j < len(oldContentLines); j++ {
			if originalLines[i+j] != oldContentLines[j] {
				isMatch = false
				break
			}
		}
		if isMatch {
			matchCount++
			matchStartIndex = i
			matchEndIndex = i + len(oldContentLines)
		}
	}

	if matchCount == 0 {
		return "", fmt.Errorf("no matching text found in the file, please carefully verify the file content and indentation")
	}

	if matchCount > 1 {
		return "", fmt.Errorf("found %d similar codes, please provide more context lines of code to precisely locate", matchCount)
	}

	var newContentLines []string
	newContentLines = append(newContentLines, original[:matchStartIndex])
	newContentLines = append(newContentLines, newContent)
	newContentLines = append(newContentLines, original[matchEndIndex:])

	return strings.Join(newContentLines, "\n"), nil
}

func NewEditFileTool(workingDir string) *EditFileTool {
	return &EditFileTool{
		workingDir: workingDir,
	}
}
