package hu60

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	config Config
}

type Config struct {
	ApiURL     string
	HTTPClient *http.Client
	XFFHeader  string
}

type CommonRequest struct {
	XFFIP string
}

type Result struct {
	Success bool
	Notice  string
}

func NewClient(apiURL string) *Client {
	return NewClientWithConfig(Config{
		ApiURL:     apiURL,
		HTTPClient: http.DefaultClient,
		XFFHeader:  "X-Forwarded-For",
	})
}
func NewClientWithConfig(config Config) *Client {
	return &Client{
		config: config,
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("too many requests! retry after: %s", res.Header.Get("Retry-After"))
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error, status code: %d, message: %s", res.StatusCode, string(b))
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(v); err != nil {
			return fmt.Errorf("invalid json format: %s", err.Error())
		}
	}

	return nil
}

func (c *Client) fullURL(suffix string) string {
	return fmt.Sprintf("%s/q.php/%s", c.config.ApiURL, suffix)
}
