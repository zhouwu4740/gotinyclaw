package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/zhouwu4740/gotinyclaw/internal/engine"
)

// FeishuBot is a bot for Feishu
type FeishuBot struct {
	client    *lark.Client
	appID     string
	appSecret string
	engine    *engine.AgentEngine
}

func NewFeishuBot(eng *engine.AgentEngine) *FeishuBot {
	appID := os.Getenv("FEISHU_APP_ID")
	appSecret := os.Getenv("FEISHU_APP_SECRET")

	if appID == "" || appSecret == "" {
		log.Fatal("FEISHU_APP_ID and FEISHU_APP_SECRET cann't empty")
	}

	client := lark.NewClient(appID, appSecret)

	return &FeishuBot{
		client:    client,
		appID:     appID,
		appSecret: appSecret,
		engine:    eng,
	}
}

func (b *FeishuBot) GetEventDispatcher() *dispatcher.EventDispatcher {
	encryptKey := os.Getenv("FEISHU_ENCRYPT_KEY")
	verifyToken := os.Getenv("FEISHU_VERIFY_TOKEN")

	handler := dispatcher.NewEventDispatcher(verifyToken, encryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			contentStr := *event.Event.Message.Content
			contentStr = strings.TrimPrefix(contentStr, `{"text":"`)
			contentStr = strings.TrimSuffix(contentStr, `"}`)

			chatId := *event.Event.Message.ChatId
			go b.handleAgentRun(chatId, contentStr)
			return nil
		}).OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
		return nil
	}).OnP2ChatAccessEventBotP2pChatEnteredV1(func(ctx context.Context, event *larkim.P2ChatAccessEventBotP2pChatEnteredV1) error {
		reporter := &FeishuReporter{
			client: b.client,
			chatId: *event.Event.ChatId,
		}
		go reporter.sendMsg("你好！发送消息即可开始对话。")
		return nil
	})

	return handler
}

func (b *FeishuBot) handleAgentRun(chatId string, prompt string) {
	reporter := &FeishuReporter{
		client: b.client,
		chatId: chatId,
	}

	err := b.engine.Run(context.Background(), prompt, reporter)

	if err != nil {
		reporter.sendMsg(fmt.Sprintf("AgentEngine running failed: %v", err))
	}
}

type FeishuReporter struct {
	client *lark.Client
	chatId string
}

func (f *FeishuReporter) sendMsg(text string) {
	textContent := map[string]string{
		"text": text,
	}

	contentBytes, _ := json.Marshal(textContent)
	contentStr := string(contentBytes)

	body := larkim.NewCreateMessageReqBodyBuilder().
		ReceiveId(f.chatId).
		MsgType(larkim.MsgTypeText).
		Content(contentStr).
		Build()

	msgReq := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.CreateMessageV1ReceiveIDTypeChatId).
		Body(body).
		Build()

	resp, err := f.client.Im.Message.Create(context.Background(), msgReq)
	if err != nil {
		log.Printf("sending feishu message failed: %v", err)
	} else if !resp.Success() {
		log.Printf("sending feishu message failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}
}

// OnMessage implements [engine.Reporter].
func (f *FeishuReporter) OnMessage(ctx context.Context, content string) {
	f.sendMsg(content)
}

// OnThinking implements [engine.Reporter].
func (f *FeishuReporter) OnThinking(ctx context.Context) {
	f.sendMsg("Model thinking ...")
}

// OnThought implements [engine.Reporter].
func (f *FeishuReporter) OnThought(ctx context.Context, thought string) {
	f.sendMsg(thought)
}

// OnToolCall implements [engine.Reporter].
func (f *FeishuReporter) OnToolCall(ctx context.Context, toolName string, args string) {
	f.sendMsg(fmt.Sprintf("call tool %s(%s)", toolName, args))
}

// OnToolResult implements [engine.Reporter].
func (f *FeishuReporter) OnToolResult(ctx context.Context, toolName string, result string, isError bool) {
	f.sendMsg(fmt.Sprintf("tool execute result: %s", result))
}
