package convo

import "time"

type ConversationOptions struct {
	Hu60Username         string
	Hu60Password         string
	Hu60APIURL           string
	Hu60WSURL            string
	OpenaiToken          string
	OpenaiModel          string
	OpenaiAPIURL         string
	OpenaiRequestTimeout time.Duration
	ConversationWindow   time.Duration
	ConversationBgFile   string
}
