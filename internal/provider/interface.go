package provider

import (
	"context"

	"github.com/zhouwu4740/gotinyclaw/internal/schema"
)

type LLMProvider interface {
	Generate(ctx context.Context, messages []schema.Message, availableTools []schema.ToolDefinition) (*schema.Message, error)
}
