package hu60

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
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
	Tag   *string
}

type ListMsgResponse struct {
	MsgCount int   `json:"msgCount"`
	MsgList  []Msg `json:"msgList"`
}

type ListMsgOptions struct {
	PageSize int
}

func (c *Client) ListMsg(ctx context.Context, sid string, opts ListMsgOptions) (response ListMsgResponse, err error) {
	suffix := fmt.Sprintf("/%s/msg.index.@.no.json?_content=json", sid)
	if opts.PageSize > 0 {
		suffix += fmt.Sprintf("&pageSize=%d", opts.PageSize)
	}
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
		resp, err := c.ListMsg(ctx, sid, ListMsgOptions{PageSize: 64})
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		for i, j := 0, len(resp.MsgList)-1; i < j; i, j = i+1, j-1 {
			resp.MsgList[i], resp.MsgList[j] = resp.MsgList[j], resp.MsgList[i]
		}
		for _, msg := range resp.MsgList {
			handler(msg)
		}
	}
}
