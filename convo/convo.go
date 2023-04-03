package convo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/gorilla/websocket"
	"github.com/rkonfj/hu60bot/pkg/hu60"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ConversationManager struct {
	hu60Client             *hu60.Client
	openaiClient           *openai.Client
	options                ConversationOptions
	conversationContext    cache.Cache[string, []openai.ChatCompletionMessage]
	openaiTokensUsage      openai.Usage
	botReadyTask           sync.WaitGroup
	botUid                 int
	botSid                 string
	botMsgChan             chan hu60.Msg
	conversationBackground []string
}

type SeWsEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
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
	openaiCfg := openai.DefaultConfig(options.OpenaiToken)
	openaiCfg.BaseURL = options.OpenaiAPIURL

	bg := []string{}
	file, err := os.Open(options.ConversationBgFile)
	if err == nil {
		defer file.Close()
		logrus.Info("bot loading conversation background file: ", options.ConversationBgFile)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			bg = append(bg, scanner.Text())
			logrus.Debug("bot loaded words: ", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			logrus.Warn(err)
		}
	}

	cm := ConversationManager{
		hu60Client:   hu60.NewClient(options.Hu60APIURL),
		openaiClient: openai.NewClientWithConfig(openaiCfg),
		options:      options,
		conversationContext: cache.NewCache[string, []openai.ChatCompletionMessage]().
			WithMaxKeys(4096).WithTTL(time.Second * 30),
		botReadyTask:           sync.WaitGroup{},
		botMsgChan:             make(chan hu60.Msg, 64),
		conversationBackground: bg,
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
	cm.processMissedMsgs()
	cm.startMsgsFetcher()
	for msg := range cm.botMsgChan {
		logrus.Debug("watched msg: ", msg)
		if msg.Content[0].Type != "atMsg" {
			logrus.Warn("skip non @ msg")
			continue
		}
		conversationKey := fmt.Sprintf("%d", msg.ByUID)

		content := getTextMessage(msg.Content)

		answer, isNewConvo, err := cm.Ask(content, conversationKey)
		if err != nil {
			logrus.Error("askAI error: ", err.Error())
			answerHu60(cm.hu60Client, resp.Sid, msg, err.Error(), isNewConvo)
			continue
		}
		answerHu60(cm.hu60Client, resp.Sid, msg, answer, isNewConvo)
	}
}

func (cm *ConversationManager) startMsgsFetcher() {
	go func() {
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
	}()
}

func (cm *ConversationManager) processMissedMsgs() {
	resp, err := cm.hu60Client.ListMsg(context.Background(), cm.botSid, hu60.ListMsgOptions{PageSize: 64})
	if err != nil {
		logrus.Warnf("missed msgs process failed. (%w)", err)
		return
	}
	logrus.Debugf("bot processed %d missed msgs", len(resp.MsgList))
	for i, j := 0, len(resp.MsgList)-1; i < j; i, j = i+1, j-1 {
		resp.MsgList[i], resp.MsgList[j] = resp.MsgList[j], resp.MsgList[i]
	}
	for _, msg := range resp.MsgList {
		cm.botMsgChan <- msg
	}
}

func (cm *ConversationManager) startTheMarkMsgReadTask() {
	go func() {
		for {
			time.Sleep(time.Minute)
			r, err := cm.hu60Client.SetMsgIsRead(context.Background(), cm.botSid, 1)
			if err != nil {
				logrus.Trace("hu60.setMsgIsRead error: ", err)
			}
			logrus.Trace("hu60.setMsgIsRead response: ", r)
		}
	}()
	go func() {
		signals := make(chan os.Signal, 1)
		defer close(signals)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		for sig := range signals {
			r, _ := cm.hu60Client.SetMsgIsRead(context.Background(), cm.botSid, 1)
			logrus.Infof("signal: %s, mark all at msgs is read (%d msgs). done", sig, r.Result.Update)
			os.Exit(0)
		}
	}()
}

func (cm *ConversationManager) startHeartbeatTask(conn *websocket.Conn) {
	go func(c *websocket.Conn) {
		for {
			time.Sleep(65 * time.Second)
			logrus.Trace(`bot ping... {"action": "ping"}`)
			if err := c.WriteMessage(websocket.TextMessage, []byte(`{"action": "ping"}`)); err != nil {
				c.Close()
				break
			}
		}
	}(conn)
}

func (cm *ConversationManager) connectWs() error {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s?_sid=%s", cm.options.Hu60WSURL, cm.botSid), nil)
	if err == nil {
		logrus.Infof("bot connect websocket (%s) succeed", cm.options.Hu60WSURL)
		logrus.Infof("bot watching for chat now. sid is %s, conversation window is %s",
			cm.botSid, cm.options.ConversationWindow.String())
		cm.startHeartbeatTask(conn)
		cm.startTheMarkMsgReadTask()
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Error("websocket error: ", err)
				time.Sleep(5 * time.Second)
				return cm.connectWs()
			}
			var seEvent SeWsEvent
			err = json.Unmarshal(message, &seEvent)
			if err != nil {
				logrus.Error("connectWs.ReadMessage: invalid hu60wap6 ws event format: ", err)
				continue
			}
			if seEvent.Event != "msg" {
				if seEvent.Event == "ping" {
					logrus.Trace("bot pong ", string(message))
				}
				continue
			}
			var event WsEvent
			err = json.Unmarshal(message, &event)
			if err != nil {
				logrus.Error("connectWs.ReadMessage: invalid hu60wap6 ws event format: ", err)
				continue
			}
			if event.Data.Type != 1 {
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
	logrus.Debugf("convoKey: %s, askAI request: %s", conversationKey, words)
	var conversationMsgs []openai.ChatCompletionMessage

	if msgs, ok := cm.conversationContext.Get(conversationKey); ok {
		conversationMsgs = msgs
	} else {
		isNewConvo = true
		conversationMsgs = []openai.ChatCompletionMessage{}
		for _, syswords := range cm.conversationBackground {
			conversationMsgs = append(conversationMsgs, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: syswords,
			})
		}
	}

	conversationMsgs = append(conversationMsgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: words,
	})

	answer, usage, err := askAI(cm.openaiClient, cm.options.OpenaiModel, cm.options.OpenaiRequestTimeout, conversationMsgs)
	if err != nil {
		if strings.Contains(err.Error(), "reduce the length") {
			err = fmt.Errorf("答案丢失！会话长度触限，下一次对话将开启新会话。(%s)", err.Error())
			cm.MarkExpired(conversationKey)
		}
		return
	}
	logrus.Debugf("convoKey: %s, askAI response: %s", conversationKey, answer)

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
