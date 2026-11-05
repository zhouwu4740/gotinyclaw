// cmd/claw/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	"github.com/zhouwu4740/gotinyclaw/internal/config"
	"github.com/zhouwu4740/gotinyclaw/internal/engine"
	"github.com/zhouwu4740/gotinyclaw/internal/feishu"
	"github.com/zhouwu4740/gotinyclaw/internal/provider"
	"github.com/zhouwu4740/gotinyclaw/internal/tools"
)

func main() {

	// 0. 优先加载 .env 文件中的环境变量（不覆盖已存在的系统环境变量）
	if err := config.LoadDotEnv(); err != nil {
		log.Fatalf("load .env failed: %v", err)
	}

	workDir, _ := os.Getwd()

	// 1. 初始化真实的 Provider大脑 (指向智谱 GLM-4.5)
	// 这里你可以任意切换 NewZhipuClaudeProvider 或 NewZhipuOpenAIProvider，效果完全一致！
	// LLM 配置优先从环境变量读取，缺省时回落到本地默认值
	llmAPIKey := os.Getenv("LLM_API_KEY")
	llmBaseURL := os.Getenv("LLM_BASE_URL")
	if llmBaseURL == "" {
		llmBaseURL = "http://localhost:11434/v1"
	}
	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		llmModel = "qwen3.8:27b-mlx"
	}
	llmProvider := provider.NewOpenAIProvider(llmAPIKey, llmBaseURL, llmModel)

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
	// prompt := "将 scores.txt 文件中王五的分数修改为 99, 并反馈改前的分数是多少"
	// prompt := "并发读取docs目录下的所有文件，将内容写入该目录下的 all.txt 文件"

	bot := feishu.NewFeishuBot(eng)

	handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())

	http.HandleFunc("/webhook/event", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[webhook] received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		handler(w, r)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "48080"
	}
	log.Printf("Feishu server started, listening on :%s, webhook path: /webhook/event\n", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server start fail: %v", err)
	}
}
