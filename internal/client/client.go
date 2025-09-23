package client

import (
	"net/http"
	"time"
)

// Client represents client for server interaction
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates new client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}
