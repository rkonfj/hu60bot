package hu60

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ReplyTopicRequest struct {
	CommonRequest
	Token   string
	Content string
	TopicID int
}

type ReplyTopicResponse struct {
	Result
	Url string
}

type GetTopicResponse struct {
	Token string
}

func (c *Client) ReplyTopic(ctx context.Context, sid string, request ReplyTopicRequest) (response ReplyTopicResponse, err error) {
	suffix := fmt.Sprintf("/bbs.newreply.%d.json", request.TopicID)
	requestBody := fmt.Sprintf("token=%s&content=%s&go=1", request.Token, url.QueryEscape(request.Content))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(suffix), strings.NewReader(requestBody))
	if err != nil {
		return
	}
	if request.XFFIP != "" {
		req.Header.Set(c.config.XFFHeader, request.XFFIP)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "hu60_sid="+sid)
	err = c.sendRequest(req, &response)

	if !response.Success {
		err = errors.New(response.Notice)
	}
	return
}

func (c *Client) GetTopic(ctx context.Context, topicid int, sid string) (response GetTopicResponse, err error) {
	suffix := fmt.Sprintf("/bbs.topic.%d.json", topicid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.fullURL(suffix), nil)
	if err != nil {
		return
	}
	req.Header.Set("Cookie", "hu60_sid="+sid)
	err = c.sendRequest(req, &response)
	return
}
