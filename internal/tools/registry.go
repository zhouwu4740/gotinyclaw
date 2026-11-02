package tools

import (
	"context"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type Registry interface {
	GetAvailableTools() []schema.ToolDefinition
	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
}
