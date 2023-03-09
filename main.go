package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/rkonfj/hu60bot/pkg/hu60"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	openai "github.com/sashabaranov/go-openai"
)

var (
	username              string
	password              string
	hu60ApiURL            string
	openAIToken           string
	conversationWindowStr string
	conversationWindow    time.Duration
	conversationContext   cache.Cache[string, []openai.ChatCompletionMessage] = cache.NewCache[string, []openai.ChatCompletionMessage]().
				WithMaxKeys(4096).WithTTL(time.Second * 30)
	logLevel string
)

func main() {
	var cmd = &cobra.Command{
		Use:    "hu60bot",
		Short:  "A hu60wap6 robot",
		Args:   cobra.NoArgs,
		PreRun: programInit,
		Run:    listen,
	}

	cmd.Flags().StringVar(&conversationWindowStr, "conversation-window", "30m", "conversation valid time. example: 1m, 1h, 1d ...")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "logging component's log level. example: error, warn, info, debug ...")
	cmd.Flags().StringVarP(&username, "hu60user", "u", "", "robot username for login hu60wap6")
	cmd.Flags().StringVarP(&password, "hu60pass", "p", "", "robot password for login hu60wap6")
	cmd.Flags().StringVar(&hu60ApiURL, "hu60api", "https://hu60.cn", "hu60wap6's API URL")
	cmd.Flags().StringVar(&openAIToken, "openai-token", "", "token for access OpenAI's API")

	cmd.MarkFlagRequired("hu60user")
	cmd.MarkFlagRequired("hu60pass")
	cmd.MarkFlagRequired("openai-token")

	cmd.Execute()
}

func programInit(cmd *cobra.Command, args []string) {
	if conversationWindowDuration, err := time.ParseDuration(conversationWindowStr); err == nil {
		conversationWindow = conversationWindowDuration
	} else {
		panic(err)
	}
	ll, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Debug("you can see debug level log now")
	updateConversationStats()
}

func listen(cmd *cobra.Command, args []string) {
	client := hu60.NewClient(hu60ApiURL)
	openAIClient := openai.NewClient(openAIToken)
	resp, err := client.Login(
		context.Background(),
		hu60.LoginRequest{Username: username, Password: password},
	)
	if err != nil {
		panic(err)
	}

	logrus.Info("your sid is ", resp.Sid, ", conversation window is ", conversationWindowStr, ", watching for chat now")

	client.WatchMsg(context.Background(), resp.Sid, func(msg hu60.Msg) {
		logrus.Debug("watched msg: ", msg)
		if msg.Content[0].Type != "atMsg" {
			logrus.Warn("skip non @ msg")
			return
		}
		conversationKey := fmt.Sprintf("%d", msg.ByUID)

		var (
			conversationMsgs []openai.ChatCompletionMessage
			newConversation  bool
		)
		if msgs, ok := conversationContext.Get(conversationKey); ok {
			conversationMsgs = msgs
		} else {
			newConversation = true
			conversationMsgs = []openai.ChatCompletionMessage{}
		}

		content := getTextMessage(msg.Content)
		logrus.Debug("getTextMessage: ", content)

		openaiMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		}

		conversationMsgs = append(conversationMsgs, openaiMsg)

		responseText, err := askAI(openAIClient, conversationMsgs)
		if err != nil {
			logrus.Error("askAI error: ", err.Error())
			answerHu60(client, resp.Sid, msg, err.Error(), newConversation)
			return
		}

		logrus.Debug("askAI response: ", responseText)

		conversationMsgs = append(conversationMsgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: responseText,
		})

		conversationContext.Set(conversationKey, conversationMsgs, conversationWindow)
		answerHu60(client, resp.Sid, msg, responseText, newConversation)
		updateConversationStats()
	})

}

func updateConversationStats() {
	stats := fmt.Sprintf(`Total Conversation: %d, Invalid Conversation: %d`, conversationContext.Stat().Added, conversationContext.Stat().Evicted)
	stats += "\n\n"
	stats += strings.Join(conversationContext.Keys(), "\n")
	stats += "\n"
	err := os.WriteFile("conversation.stat.txt", []byte(stats), 0644)
	if err != nil {
		logrus.Error("update ConversationStats error: ", err.Error())
	}
}
