package engine

import (
	"context"
)

// Reporter is an interface for reporting agent events.
type Reporter interface {
	// OnThinking is called when the agent is thinking.
	OnThinking(ctx context.Context)
	// OnThought is called when the agent has a thought.
	OnThought(ctx context.Context, thought string)
	// OnToolCall is called when the agent calls a tool.
	OnToolCall(ctx context.Context, toolName string, args string)
	// OnToolResult is called when the tool return result.
	OnToolResult(ctx context.Context, toolName string, result string, isError bool)
	// OnMessage is called when the model declares a final text answer to the user.
	OnMessage(ctx context.Context, content string)
}
