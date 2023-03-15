package convo

import "time"

type ConversationOptions struct {
	Hu60Username       string
	Hu60Password       string
	Hu60APIURL         string
	OpenaiToken        string
	OpenaiModel        string
	ConversationWindow time.Duration
}
