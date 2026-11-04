package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type BashTool struct {
	workDir string
}

// Name implements [BaseTool].
func (b *BashTool) Name() string {
	return "bash"
}

// Definition implements [BaseTool].
func (b *BashTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "bash",
		Description: "Run a bash command in the working directory. The command is passed as a string.",
		InputSchema: map[string]any{
			"type": "string",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The bash command to run.",
				},
			},
			"required": []string{"command"},
		},
	}
}

type bashArgs struct {
	Command string `json:"command"`
}

// Execute implements [BaseTool].
func (b *BashTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {

	// 1. Parse the input arguments.
	var input bashArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("parsing bash command arguments failed: %w", err)
	}

	// 2. Set a timeout for the command.
	timeout := 30 * time.Second
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// 3. Execute the command.
	cmd := exec.CommandContext(timeoutContext, "bash", "-c", input.Command)
	// 4. Set the working directory for the command.
	cmd.Dir = b.workDir
	// 5. Execute the command and capture the output.
	log.Printf("Executing command: %s", input.Command)
	output, err := cmd.CombinedOutput()
	// 6. Check if the command failed.
	if err != nil {
		return "", fmt.Errorf("bash command failed: %w", err)
	}
	outputStr := string(output)
	// 7. Check if the command timed out.
	if timeoutContext.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("bash command timed out, output: %s, timeout: %s", outputStr, timeout)
	}

	if outputStr == "" {
		return "command executed successfully and no output", nil
	}

	return fmt.Sprintf("Bash command output: %s\n", outputStr), nil
}

func NewBashTool(workdir string) *BashTool {
	return &BashTool{workDir: workdir}
}
