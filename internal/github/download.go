package github

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// WorkflowDownloader defines the interface for downloading workflow files.
type WorkflowDownloader interface {
	DownloadWorkflow(url string) ([]byte, error)
}

// DownloadWorkflow downloads a GitHub Actions workflow YAML file from the given full URL or local file path.
func (c *Client) DownloadWorkflow(url string) ([]byte, error) {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		slog.Debug("Downloading workflow from URL", "url", url)

		if strings.Contains(url, "github.com/") && strings.Contains(url, "/blob/") {
			url = convertToRawURL(url)
		}
		// Default: HTTP(S) download
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

	slog.Debug("Downloading workflow from local file", "url", url)

	// Support local files: file:// or plain path
	path := url
	if strings.HasPrefix(url, "file://") {
		path = url[7:]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read local file: %w", err)
	}
	return data, nil
}

// convertToRawURL converts a URL of the form github.com/{owner}/{repo}/blob/{branch}/{path}
// to raw.githubusercontent.com/{owner}/{repo}/{branch}/{path}.
func convertToRawURL(url string) string {
	if strings.HasPrefix(url, "https://github.com/") {

		slog.Debug("Converting GitHub URL to raw URL", "url", url)

		url = strings.Replace(url, "https://github.com/", "https://raw.githubusercontent.com/", 1)
		url = strings.Replace(url, "/blob/", "/", 1)
		return url
	}
	return url
}
