package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) request(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s%s", c.BaseURL, path),
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.APIKey)
	return c.HTTPClient.Do(req)
}

func (c *Client) requestMultipart(method, path string, body *bytes.Buffer, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", contentType)

	return c.HTTPClient.Do(req)
}
