package hu60

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type LoginRequest struct {
	CommonRequest
	Username string
	Password string
}

type LoginResponse struct {
	Result
	Uid int
	Sid string
}

type GetProfileRequest struct {
	CommonRequest
	Sid string
}

type GetProfileResponse struct {
	Uid         int
	Name        string
	Mail        string
	Regtime     int64
	HasRegPhone bool
	SiteAdmin   bool
}

func (c *Client) Login(ctx context.Context, request LoginRequest) (respone LoginResponse, err error) {
	requestBody := fmt.Sprintf("type=1&name=%s&pass=%s&go=1", request.Username, request.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL("/user.login.json"), strings.NewReader(requestBody))
	if err != nil {
		return
	}

	if request.XFFIP != "" {
		req.Header.Set(c.config.XFFHeader, request.XFFIP)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = c.sendRequest(req, &respone)

	if !respone.Success {
		err = errors.New(respone.Notice)
		return
	}
	return
}

func (c *Client) GetProfile(ctx context.Context, request GetProfileRequest) (response GetProfileResponse, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.fullURL("/user.index.json"), nil)
	if err != nil {
		return
	}
	if request.XFFIP != "" {
		req.Header.Set(c.config.XFFHeader, request.XFFIP)
	}
	req.Header.Set("Cookie", "hu60_sid="+request.Sid)
	err = c.sendRequest(req, &response)
	if err != nil {
		return
	}
	if response.Uid == 0 {
		err = fmt.Errorf("can not get user uid. sid %s is not correct", request.Sid)
	}
	return
}
