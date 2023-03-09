package hu60

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Msg struct {
	ID      int `json:"id"`
	ToUID   int `json:"touid"`
	ByUID   int `json:"byuid"`
	Type    int
	Read    int `json:"isread"`
	Content []MsgContent
}

type MsgContent struct {
	Type    string
	UID     int `json:"uid"`
	Pos     string
	URL     string    `json:"url"`
	MsgUnit []MsgUnit `json:"msg"`
}

type MsgUnit struct {
	Type  string
	Value *string
	Data  *string
}

type ListMsgResponse struct {
	MsgCount int   `json:"msgCount"`
	MsgList  []Msg `json:"msgList"`
}

func (c *Client) ListMsg(ctx context.Context, sid string) (response ListMsgResponse, err error) {
	suffix := fmt.Sprintf("/%s/msg.index.@.no.json?_content=json", sid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.fullURL(suffix), nil)
	if err != nil {
		return
	}
	err = c.sendRequest(req, &response)
	return
}

func (c *Client) WatchMsg(ctx context.Context, sid string, handler func(msg Msg)) {
	for {
		time.Sleep(10 * time.Second)
		resp, err := c.ListMsg(ctx, sid)
		if err != nil {
			fmt.Println("[ERR]", err.Error())
			continue
		}
		for _, msg := range resp.MsgList {
			handler(msg)
		}
	}
}
