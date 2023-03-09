package hu60

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Result
	Uid int
	Sid string
}

func (c *Client) Login(ctx context.Context, request LoginRequest) (respone LoginResponse, err error) {
	requestBody := fmt.Sprintf("type=1&name=%s&pass=%s&go=1", request.Username, request.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL("/user.login.json"), strings.NewReader(requestBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = c.sendRequest(req, &respone)

	if !respone.Success {
		err = errors.New(respone.Notice)
		return
	}
	return
}
