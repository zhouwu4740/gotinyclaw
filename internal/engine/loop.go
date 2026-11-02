package engine

import (
	"context"
	"log"

	"github.com/zhouwu4740/gotinyclaw/internal/provider"
	"github.com/zhouwu4740/gotinyclaw/internal/schema"
	"github.com/zhouwu4740/gotinyclaw/internal/tools"
)

type AgentEngine struct {
	provider provider.LLMProvider
	registry tools.Registry

	// 工作区
	WorkDir string
}

func NewAgentEngine(provider provider.LLMProvider, registry tools.Registry, workDir string) *AgentEngine {
	return &AgentEngine{
		provider: provider,
		registry: registry,
		WorkDir:  workDir,
	}
}

func (e *AgentEngine) Run(ctx context.Context, userPrompt string) error {
	log.Printf("Received user prompt: %s", userPrompt)

	contextHistory := []schema.Message{
		{Role: schema.RoleSystem, Content: "You are a helpful assistant and have full access to tools in the workspace."},
		{Role: schema.RoleUser, Content: userPrompt},
	}

	turnCount := 0
	for {
		turnCount++
		// 获取可以使用的工具集
		availableTools := e.registry.GetAvailableTools()

		response, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return err
		}

		contextHistory = append(contextHistory, *response)

		if response.Content != "" {
			log.Printf("Assistant response: %s", response.Content)
		}

		if len(response.ToolCalls) == 0 {
			log.Println("[Engine] No tool calls in response, ending conversation.")
			break
		}

		log.Printf("[Engine] Model request to execute %d tools", len(response.ToolCalls))

		for _, toolCall := range response.ToolCalls {
			result := e.registry.Execute(ctx, toolCall)
			if result.IsError {
				log.Printf("[Engine] Tool %s returned an error: %s", toolCall.Name, result.Output)
			} else {
				log.Printf("[Engine] Tool %s returned: %s", toolCall.Name, result.Output)
			}

			observationMsg := schema.Message{
				Role:       schema.RoleTool,
				Content:    result.Output,
				ToolCallID: toolCall.ID,
			}
			contextHistory = append(contextHistory, observationMsg)
		}

	}
	return nil
}
