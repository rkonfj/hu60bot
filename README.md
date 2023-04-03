# Introducing hu60bot

### Usage of bot 
```
A hu60wap6 robot

Usage:
  hu60bot [flags]

Flags:
      --conversation-bg string       each line is decorated as a system role message sent to openai (default "bg.txt")
      --conversation-window string   conversation valid time. example: 1m, 1h, 1d ... (default "30m")
  -h, --help                         help for hu60bot
      --hu60api string               hu60wap6's api url (default "https://hu60.cn")
  -p, --hu60pass string              robot password for login hu60wap6
  -u, --hu60user string              robot username for login hu60wap6
      --hu60ws string                hu60wap6's websocket endpoint url (default "wss://hu60.cn/ws/msg")
      --log-level string             logging level. example: error, warn, info, debug ... (default "info")
      --openai-api string            openai's api url with version (default "https://api.openai.com/v1")
      --openai-model string          id of the openai model to use. https://platform.openai.com/docs/models/overview (default "gpt-3.5-turbo")
      --openai-timeout string        timeout for requesting openai api (default "65s")
  -k, --openai-token string          api key for access openai. https://platform.openai.com/account/api-keys
```

### Usage of websocket server 
```
A hu60wap6 websocket server

Usage:
  hu60ws [flags]

Flags:
      --canal-client-destination string   canal client destination for watching hu60wap6 db (default "hu60bot")
      --canal-host string                 canal host for watching hu60wap6 db (default "127.0.0.1")
      --canal-port int                    canal port for watching hu60wap6 db (default 11111)
      --disable-action strings            websocket server disabled bot action. can be specified multiple times
  -h, --help                              help for hu60ws
      --hu60api string                    hu60wap6's api url (default "https://hu60.cn")
  -l, --listen string                     websocket server listen address (default "127.0.0.1:4860")
      --log-level string                  logging level. example: error, warn, info, debug ... (default "info")
      --wspu int                          websocket server connections limit per user (default 10)
      --xff string                        header will be sent to hu60api which value is the client's original ip (default "X-Forwarded-For")
```