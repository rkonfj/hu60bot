package main

import (
	"sync"

	"github.com/rkonfj/hu60bot/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "hu60ws",
		Short:   "A hu60wap6 websocket server",
		Args:    cobra.NoArgs,
		PreRunE: wsInit,
		RunE:    wsAction,
	}
	cmd.Flags().String("log-level", "info", "logging level. example: error, warn, info, debug ...")

	cmd.Flags().String("hu60api", "https://hu60.cn", "hu60wap6's api url")

	cmd.Flags().String("canal-host", "127.0.0.1", "canal host for watching hu60wap6 db")
	cmd.Flags().Int("canal-port", 11111, "canal port for watching hu60wap6 db")
	cmd.Flags().String("canal-client-destination", "hu60bot", "canal client destination for watching hu60wap6 db")

	cmd.Flags().StringSlice("disable-action", []string{}, "websocket server disabled bot action. can be specified multiple times")

	cmd.Flags().StringP("listen", "l", "127.0.0.1:4860", "websocket server listen address")
	cmd.Flags().Int("wspu", 10, "websocket server connections limit per user")

	cmd.Flags().String("xff", "X-Forwarded-For", "header will be sent to hu60api which value is the client's original ip")

	cmd.Execute()
}

func wsInit(cmd *cobra.Command, args []string) error {
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

func wsAction(cmd *cobra.Command, args []string) error {
	serverOpts, err := processServerOptions(cmd)
	if err != nil {
		return err
	}

	canalOpts, err := processCanalOptions(cmd)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	websocketManager := server.NewWebsocketManager(serverOpts)
	canalManager := server.NewCanalManager(canalOpts, websocketManager)

	go func() {
		err := websocketManager.Run()
		if err != nil {
			logrus.Errorf("websocket server is disabled. (%s)", err.Error())
		}
		wg.Done()
	}()

	go func() {
		err := canalManager.Run()
		if err != nil {
			logrus.Errorf("canal is disabled. (%s)", err.Error())
		}
		websocketManager.OnCanalStartFailed()
		wg.Done()
	}()
	wg.Wait()
	return nil
}

func processServerOptions(cmd *cobra.Command) (options server.ServerOptions, err error) {
	options.Hu60wap6APIURL, err = cmd.Flags().GetString("hu60api")
	if err != nil {
		return
	}
	options.BotXFF, err = cmd.Flags().GetString("xff")
	if err != nil {
		return
	}
	options.DisabledActions, err = cmd.Flags().GetStringSlice("disable-action")
	if err != nil {
		return
	}
	options.ConnectionLimitPerUser, err = cmd.Flags().GetInt("wspu")
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
