package convo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/gorilla/websocket"
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
	botReadyTask        sync.WaitGroup
	botUid              int
	botSid              string
	botMsgChan          chan hu60.Msg
}

type WsEvent struct {
	Event string    `json:"event"`
	Data  WsHu60Msg `json:"data"`
}

type WsHu60Msg struct {
	ID      int    `json:"id"`
	ByUID   int    `json:"byuid"`
	ToUID   int    `json:"touid"`
	Type    int    `json:"type"`
	Read    int    `json:"isread"`
	Content string `json:"content"`
}

func NewConversationManager(options ConversationOptions) *ConversationManager {
	cm := ConversationManager{
		hu60Client:   hu60.NewClient(options.Hu60APIURL),
		openaiClient: openai.NewClient(options.OpenaiToken),
		options:      options,
		conversationContext: cache.NewCache[string, []openai.ChatCompletionMessage]().
			WithMaxKeys(4096).WithTTL(time.Second * 30),
		botReadyTask: sync.WaitGroup{},
		botMsgChan:   make(chan hu60.Msg),
	}
	cm.updateConversationStats()
	cm.botReadyTask.Add(1)
	return &cm
}

func (cm *ConversationManager) Run() {
	defer close(cm.botMsgChan)

	resp, err := cm.hu60Client.Login(
		context.Background(),
		hu60.LoginRequest{Username: cm.options.Hu60Username, Password: cm.options.Hu60Password},
	)
	if err != nil {
		logrus.Fatal(err)
	}

	cm.botSid = resp.Sid
	cm.botUid = resp.Uid
	cm.botReadyTask.Done()
	for msg := range cm.botMsgChan {
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
	}
}

func (cm *ConversationManager) OnCanalStartFailed() {
	// websocket
	wsurl, err := url.Parse(cm.options.Hu60WSURL)
	if err != nil {
		logrus.Fatal("invalid hu60wap6's websocket endpoint url")
	}
	cm.botReadyTask.Wait()

	err = cm.connectWs()
	if err == nil {
		return
	}

	logrus.Infof("bot connect websocket (%s) error, fallback to http loop", wsurl.String())
	logrus.Infof("bot watching for chat now. sid is %s, conversation window is %s", cm.botSid, cm.options.ConversationWindow.String())
	// http loop
	cm.hu60Client.WatchMsg(context.Background(), cm.botSid, func(msg hu60.Msg) {
		cm.botMsgChan <- msg
	})
}

func (cm *ConversationManager) OnCanalStartSucceed() {
	logrus.Infof("bot watching for chat now. conversation window is %s", cm.options.ConversationWindow.String())
}

func (cm *ConversationManager) connectWs() error {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s?_sid=%s", cm.options.Hu60WSURL, cm.botSid), nil)
	if err == nil {
		logrus.Infof("bot connect websocket (%s) succeed", cm.options.Hu60WSURL)
		logrus.Infof("bot watching for chat now. sid is %s, conversation window is %s",
			cm.botSid, cm.options.ConversationWindow.String())
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Error("websocket error: ", err)
				time.Sleep(5 * time.Second)
				return cm.connectWs()
			}
			var event WsEvent
			err = json.Unmarshal(message, &event)
			if err != nil {
				logrus.Error("connectWs.ReadMessage: invalid hu60wap6 ws event format: ", err)
				continue
			}
			if event.Event != "msg" || event.Data.Type != 1 {
				continue
			}
			var c []hu60.MsgContent
			err = json.Unmarshal([]byte(event.Data.Content), &c)
			if err != nil {
				logrus.Error("connectWs.ReadMessage: invalid hu60wap6 msg content format: ", err)
				continue
			}
			cm.botMsgChan <- hu60.Msg{
				ID:      event.Data.ID,
				ByUID:   event.Data.ByUID,
				ToUID:   event.Data.ToUID,
				Type:    event.Data.Type,
				Read:    event.Data.Read,
				Content: c,
			}
		}
	}
	return err
}

func (cm *ConversationManager) OnHu60Msg(msg hu60.Msg) {
	if msg.ToUID == cm.botUid {
		cm.botMsgChan <- msg
		return
	}
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

func (cm *ConversationManager) MarkExpired(conversationKey string) {
	cm.conversationContext.Invalidate(conversationKey)
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
