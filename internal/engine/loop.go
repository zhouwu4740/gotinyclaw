package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/zhouwu4740/gotinyclaw/internal/provider"
	"github.com/zhouwu4740/gotinyclaw/internal/schema"
	"github.com/zhouwu4740/gotinyclaw/internal/tools"
)

type AgentEngine struct {
	provider provider.LLMProvider
	registry tools.Registry

	// 工作区
	WorkDir string

	// EnableThinking
	EnableThinking bool
}

func NewAgentEngine(provider provider.LLMProvider, registry tools.Registry, workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       provider,
		registry:       registry,
		WorkDir:        workDir,
		EnableThinking: enableThinking,
	}
}

func (e *AgentEngine) Run(ctx context.Context, userPrompt string, reporter Reporter) error {
	log.Printf("Received user prompt: %s", userPrompt)

	contextHistory := []schema.Message{
		{Role: schema.RoleSystem, Content: "You are a helpful assistant and have full access to tools in the workspace."},
		{Role: schema.RoleUser, Content: userPrompt},
	}

	turnCount := 0
	for {
		turnCount++
		log.Println("[Engine] Starting turn ", turnCount)
		// 获取可以使用的工具集
		availableTools := e.registry.GetAvailableTools()

		// 慢思考
		if e.EnableThinking {
			log.Println("[Engine] Thinking...")
			// Notify the reporter that the engine is thinking
			if reporter != nil {
				reporter.OnThinking(ctx)
			}
			response, err := e.provider.Generate(ctx, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("thinking error: %w", err)
			}
			// Notify the reporter of the agent's thought
			reporter.OnThought(ctx, response.Content)

			if response.Content != "" {
				log.Printf("Assistant response: %s", response.Content)
				contextHistory = append(contextHistory, *response)
			}
		}

		response, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return err
		}

		contextHistory = append(contextHistory, *response)

		if response.Content != "" && reporter != nil {
			reporter.OnMessage(ctx, response.Content)
		}

		if len(response.ToolCalls) == 0 {
			break
		}

		observationMsgs := make([]schema.Message, len(response.ToolCalls))

		var wg sync.WaitGroup

		for i, toolCall := range response.ToolCalls {
			wg.Add(1) // Add a wait group for each tool call
			go func(idx int, call schema.ToolCall) {
				defer wg.Done()
				reporter.OnToolCall(ctx, call.Name, string(call.Arguments))
				result := e.registry.Execute(ctx, toolCall)
				if result.IsError {
					reporter.OnToolResult(ctx, call.Name, result.Output, true)
				} else {
					reporter.OnToolResult(ctx, call.Name, result.Output, false)
				}
				observationMsgs[idx] = schema.Message{
					Role:       schema.RoleTool,
					Content:    result.Output,
					ToolCallID: call.ID,
				}
			}(i, toolCall)
		}

		wg.Wait()
		contextHistory = append(contextHistory, observationMsgs...)
	}
	return nil
}
