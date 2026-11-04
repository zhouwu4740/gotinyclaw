package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type BaseTool interface {
	// Name
	Name() string
	// Definition
	Definition() schema.ToolDefinition
	// Execute
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

type Registry interface {
	Register(tool BaseTool)
	GetAvailableTools() []schema.ToolDefinition
	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
}

type registryImpl struct {
	tools map[string]BaseTool
}

func NewRegistry() Registry {
	return &registryImpl{
		tools: make(map[string]BaseTool),
	}
}

func (r *registryImpl) Register(tool BaseTool) {
	r.tools[tool.Name()] = tool
}

func (r *registryImpl) GetAvailableTools() []schema.ToolDefinition {
	definitions := []schema.ToolDefinition{}
	for _, tool := range r.tools {
		definitions = append(definitions, tool.Definition())
	}
	return definitions
}

func (r *registryImpl) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	tool, exists := r.tools[call.Name]
	if !exists {
		errMsg := fmt.Sprintf("Tool %s not found", call.Name)
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     errMsg,
			IsError:    true,
		}
	}
	output, err := tool.Execute(ctx, call.Arguments)
	if err != nil {
		errMsg := fmt.Sprintf("Tool %s execution failed: %s", call.Name, err.Error())
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     errMsg,
			IsError:    true,
		}
	}
	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     output,
		IsError:    false,
	}
}
