package main

import (
	"sync"
	"time"

	"github.com/rkonfj/hu60bot/convo"
	"github.com/rkonfj/hu60bot/server"
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
	cmd.Flags().String("log-level", "info", "logging level. example: error, warn, info, debug ...")
	cmd.Flags().String("hu60api", "https://hu60.cn", "hu60wap6's api url")
	cmd.Flags().String("hu60ws", "wss://hu60.cn/ws/msg", "hu60wap6's websocket endpoint url")
	cmd.Flags().StringP("openai-token", "k", "", "api key for access openai. https://platform.openai.com/account/api-keys")
	cmd.Flags().String("openai-model", openai.GPT3Dot5Turbo, "id of the openai model to use. https://platform.openai.com/docs/models/overview")
	cmd.Flags().String("openai-api", "https://api.openai.com/v1", "openai's api url with version")
	cmd.Flags().String("canal-host", "127.0.0.1", "canal host for watching hu60wap6 db")
	cmd.Flags().Int("canal-port", 11111, "canal port for watching hu60wap6 db")
	cmd.Flags().String("canal-client-destination", "hu60bot", "canal client destination for watching hu60wap6 db")

	cmd.Flags().StringP("hu60user", "u", "", "robot username for login hu60wap6")
	cmd.Flags().StringP("hu60pass", "p", "", "robot password for login hu60wap6")
	cmd.Flags().StringP("listen", "l", "127.0.0.1:4860", "websocket server listen address")

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

	options.Hu60APIURL, err = cmd.Flags().GetString("hu60api")
	if err != nil {
		return
	}
	options.Hu60WSURL, err = cmd.Flags().GetString("hu60ws")
	if err != nil {
		return
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

	options.Hu60Username, err = cmd.Flags().GetString("hu60user")
	if err != nil {
		return
	}
	options.Hu60Password, err = cmd.Flags().GetString("hu60pass")
	return
}

func processServerOptions(cmd *cobra.Command) (options server.ServerOptions, err error) {
	options.Hu60wap6APIURL, err = cmd.Flags().GetString("hu60api")
	if err != nil {
		return
	}
	options.Listen, err = cmd.Flags().GetString("listen")
	return
}

func processCanalOptions(cmd *cobra.Command) (options server.CanalOptions, err error) {
	options.CanalHost, err = cmd.Flags().GetString("canal-host")
	if err != nil {
		return
	}
	options.CanalPort, err = cmd.Flags().GetInt("canal-port")
	if err != nil {
		return
	}
	options.CanalClientDestination, err = cmd.Flags().GetString("canal-client-destination")
	return
}

func botAction(cmd *cobra.Command, args []string) error {
	convoOpts, err := processConversationOptions(cmd)
	if err != nil {
		return err
	}
	serverOpts, err := processServerOptions(cmd)
	if err != nil {
		return err
	}

	canalOpts, err := processCanalOptions(cmd)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)

	conversationManager := convo.NewConversationManager(convoOpts)
	websocketManager := server.NewWebsocketManager(serverOpts, conversationManager)
	canalManager := server.NewCanalManager(canalOpts, websocketManager)

	go func() {
		conversationManager.Run()
		wg.Done()
	}()

	go func() {
		err := websocketManager.Run()
		if err != nil {
			logrus.Debugf("websocket server is disabled. (%s)", err.Error())
		}
		wg.Done()
	}()

	go func() {
		err := canalManager.Run()
		if err != nil {
			logrus.Debugf("canal is disabled. (%s)", err.Error())
		}
		conversationManager.OnCanalStartFailed()
		websocketManager.OnCanalStartFailed()
		wg.Done()
	}()
	wg.Wait()
	return nil
}
