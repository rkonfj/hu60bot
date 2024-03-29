package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rkonfj/hu60bot/convo"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "hu60bot",
		Short:   "A hu60wap6 robot",
		Args:    cobra.NoArgs,
		PreRunE: botInit,
		RunE:    botAction,
	}
	cmd.Flags().String("conversation-window", "30m", "conversation valid time. example: 1m, 1h, 1d ...")
	cmd.Flags().String("conversation-bg", "bg.txt", "each line is decorated as a system role message sent to openai")
	cmd.Flags().String("log-level", "info", "logging level. example: error, warn, info, debug ...")

	cmd.Flags().String("hu60api", "https://hu60.cn", "hu60wap6's api url")
	cmd.Flags().String("hu60ws", "", "hu60wap6's websocket endpoint url (default \"<hu60api>/ws/msg\")")
	cmd.Flags().StringP("hu60user", "u", "", "robot username for login hu60wap6")
	cmd.Flags().StringP("hu60pass", "p", "", "robot password for login hu60wap6")

	cmd.Flags().StringP("openai-token", "k", "", "api key for access openai. https://platform.openai.com/account/api-keys")
	cmd.Flags().String("openai-model", openai.GPT3Dot5Turbo, "id of the openai model to use. https://platform.openai.com/docs/models/overview")
	cmd.Flags().String("openai-api", "https://api.openai.com/v1", "openai's api url with version")
	cmd.Flags().String("openai-timeout", "65s", "timeout for requesting openai api")

	cmd.MarkFlagRequired("hu60user")
	cmd.MarkFlagRequired("hu60pass")
	cmd.MarkFlagRequired("openai-token")

	cmd.Execute()
}
func botInit(cmd *cobra.Command, args []string) error {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return err
	}
	ll, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(ll)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	return nil
}

func botAction(cmd *cobra.Command, args []string) error {
	convoOpts, err := processConversationOptions(cmd)
	if err != nil {
		return err
	}

	convo.NewConversationManager(convoOpts).Run()

	return nil
}

func processConversationOptions(cmd *cobra.Command) (options convo.ConversationOptions, err error) {
	conversationWindowStr, err := cmd.Flags().GetString("conversation-window")
	if err != nil {
		return
	}

	if conversationWindowDuration, err := time.ParseDuration(conversationWindowStr); err != nil {
		return options, err
	} else {
		options.ConversationWindow = conversationWindowDuration
	}

	options.ConversationBgFile, err = cmd.Flags().GetString("conversation-bg")
	if err != nil {
		return
	}

	options.Hu60APIURL, err = cmd.Flags().GetString("hu60api")
	if err != nil {
		return
	}
	options.Hu60WSURL, err = cmd.Flags().GetString("hu60ws")
	if err != nil {
		return
	}
	if options.Hu60WSURL == "" {
		var u *url.URL
		u, err = url.ParseRequestURI(options.Hu60APIURL)
		if err != nil {
			return
		}
		options.Hu60WSURL = fmt.Sprintf("wss://%s/ws/msg", u.Host)
	}

	options.OpenaiToken, err = cmd.Flags().GetString("openai-token")
	if err != nil {
		return
	}

	options.OpenaiModel, err = cmd.Flags().GetString("openai-model")
	if err != nil {
		return
	}

	options.OpenaiAPIURL, err = cmd.Flags().GetString("openai-api")
	if err != nil {
		return
	}

	openaiRequestDurationStr, err := cmd.Flags().GetString("openai-timeout")
	if err != nil {
		return
	}

	if openaiRequestDuration, err := time.ParseDuration(openaiRequestDurationStr); err != nil {
		return options, err
	} else {
		options.OpenaiRequestTimeout = openaiRequestDuration
	}

	options.Hu60Username, err = cmd.Flags().GetString("hu60user")
	if err != nil {
		return
	}
	options.Hu60Password, err = cmd.Flags().GetString("hu60pass")
	return
}
