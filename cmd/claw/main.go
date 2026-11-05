// cmd/claw/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/zhouwu4740/gotinyclaw/internal/engine"
	"github.com/zhouwu4740/gotinyclaw/internal/provider"
	"github.com/zhouwu4740/gotinyclaw/internal/tools"
)

func main() {

	workDir, _ := os.Getwd()

	// 1. 初始化真实的 Provider大脑 (指向智谱 GLM-4.5)
	// 这里你可以任意切换 NewZhipuClaudeProvider 或 NewZhipuOpenAIProvider，效果完全一致！
	llmProvider := provider.NewOpenAIProvider("", "http://localhost:11434/v1", "qwen3.8:27b-mlx")

	// 2. 注入伪造的工具注册表
	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 3. 实例化并运行引擎，开启 EnableThinking = true (开启慢思考阶段！)
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, false)

	// 设定测试任务
	// prompt := "我想去北京跑步，帮我查查天气适合吗？"
	// prompt := "请列出当前目录结构，并统计文件和目录数量(忽略.git目录)"
	prompt := "将 scores.txt 文件中王五的分数修改为 99, 并反馈改前的分数是多少"

	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎运行崩溃: %v", err)
	}
}
