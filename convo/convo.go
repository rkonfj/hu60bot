package convo

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/rkonfj/hu60bot/pkg/hu60"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ConversationManager struct {
	hu60Client          *hu60.Client
	openaiClient        *openai.Client
	options             ConversationOptions
	conversationContext cache.Cache[string, []openai.ChatCompletionMessage]
	openaiTokensUsage   openai.Usage
}

func NewConversationManager(options ConversationOptions) *ConversationManager {
	cm := ConversationManager{
		hu60Client:   hu60.NewClient(options.Hu60APIURL),
		openaiClient: openai.NewClient(options.OpenaiToken),
		options:      options,
		conversationContext: cache.NewCache[string, []openai.ChatCompletionMessage]().
			WithMaxKeys(4096).WithTTL(time.Second * 30),
	}
	cm.updateConversationStats()
	return &cm
}

func (cm *ConversationManager) Run() {
	resp, err := cm.hu60Client.Login(
		context.Background(),
		hu60.LoginRequest{Username: cm.options.Hu60Username, Password: cm.options.Hu60Password},
	)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("sid is ", resp.Sid, ", conversation window is ", cm.options.ConversationWindow.String(), ", watching for chat now")

	cm.hu60Client.WatchMsg(context.Background(), resp.Sid, func(msg hu60.Msg) {
		logrus.Debug("watched msg: ", msg)
		if msg.Content[0].Type != "atMsg" {
			logrus.Warn("skip non @ msg")
			return
		}
		conversationKey := fmt.Sprintf("%d", msg.ByUID)

		content := getTextMessage(msg.Content)
		logrus.Debug("getTextMessage: ", content)

		answer, isNewConvo, err := cm.Ask(content, conversationKey)
		if err != nil {
			logrus.Error("askAI error: ", err.Error())
			answerHu60(cm.hu60Client, resp.Sid, msg, err.Error(), isNewConvo)
			return
		}
		answerHu60(cm.hu60Client, resp.Sid, msg, answer, isNewConvo)
	})
}

func (cm *ConversationManager) Ask(words, conversationKey string) (answer string, isNewConvo bool, err error) {
	var conversationMsgs []openai.ChatCompletionMessage

	if msgs, ok := cm.conversationContext.Get(conversationKey); ok {
		conversationMsgs = msgs
	} else {
		isNewConvo = true
		conversationMsgs = []openai.ChatCompletionMessage{}
	}

	conversationMsgs = append(conversationMsgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: words,
	})

	answer, usage, err := askAI(cm.openaiClient, cm.options.OpenaiModel, conversationMsgs)
	if err != nil {
		return
	}
	logrus.Debug("askAI response: ", answer)

	conversationMsgs = append(conversationMsgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: answer,
	})

	cm.conversationContext.Set(conversationKey, conversationMsgs, cm.options.ConversationWindow)
	cm.openaiTokensUsage.PromptTokens += usage.PromptTokens
	cm.openaiTokensUsage.CompletionTokens += usage.CompletionTokens
	cm.openaiTokensUsage.TotalTokens += usage.TotalTokens
	cm.updateConversationStats()
	return
}

func (cm *ConversationManager) updateConversationStats() {
	stats := fmt.Sprintf(`conversation: %d/%d, openai tokens usage: %d/%d/%d`,
		cm.conversationContext.Stat().Added, cm.conversationContext.Stat().Evicted,
		cm.openaiTokensUsage.PromptTokens, cm.openaiTokensUsage.CompletionTokens, cm.openaiTokensUsage.TotalTokens)
	logrus.Debug(stats)
	stats += "\n\n"
	stats += strings.Join(cm.conversationContext.Keys(), "\n")
	stats += "\n"
	err := os.WriteFile("conversation.stat.txt", []byte(stats), 0644)
	if err != nil {
		logrus.Error("update ConversationStats error: ", err.Error())
	}
}
