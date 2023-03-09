package hu60

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ReplyChatroomRequest struct {
	Token        string
	Content      string
	ChatroomName string
}
type ReplyChatroomResponse struct {
	Result
	Url string
}

type GetChatroomResponse struct {
	Token string
}

func (c *Client) ReplyChatroom(ctx context.Context, sid string, request ReplyChatroomRequest) (response ReplyChatroomResponse, err error) {
	suffix := fmt.Sprintf("/addin.chat.%s.json", request.ChatroomName)
	requestBody := fmt.Sprintf("token=%s&content=%s&go=1", request.Token, url.QueryEscape(request.Content))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(suffix), strings.NewReader(requestBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "hu60_sid="+sid)
	err = c.sendRequest(req, &response)

	if !response.Success {
		err = errors.New(response.Notice)
	}
	return
}

func (c *Client) GetChatroom(ctx context.Context, chatroomName, sid string) (response GetChatroomResponse, err error) {
	suffix := fmt.Sprintf("/addin.chat.%s.json", chatroomName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.fullURL(suffix), nil)
	if err != nil {
		return
	}
	req.Header.Set("Cookie", "hu60_sid="+sid)
	err = c.sendRequest(req, &response)
	return
}
