package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadWorkflow downloads a GitHub Actions workflow YAML file from the given full URL.
func (c *Client) DownloadWorkflow(url string) ([]byte, error) {
	req, err := c.newRequest("GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}
