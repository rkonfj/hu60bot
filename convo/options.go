package convo

import "time"

type ConversationOptions struct {
	Hu60Username       string
	Hu60Password       string
	Hu60APIURL         string
	OpenaiToken        string
	ConversationWindow time.Duration
}
