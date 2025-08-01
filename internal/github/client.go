package github

import (
	"fmt"
	"net/http"
	"strings"
)

const baseURL = "https://api.github.com"

// Client is a GitHub API client for downloading workflows and actions.
type Client struct {
	Token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client with the given token.
func NewClient(token string) *Client {
	return &Client{
		Token:      token,
		httpClient: http.DefaultClient,
	}
}

// newRequest creates a new HTTP request with the given method and path, using the baseURL unless path is a full URL.
// It adds the Authorization header if a token is set.
func (c *Client) newRequest(method, path string) (*http.Request, error) {
	url := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		url = fmt.Sprintf("%s%s", baseURL, path)
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return req, nil
}

// Do executes the HTTP request using the client's httpClient.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
