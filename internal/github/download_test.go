package github

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadWorkflow_LocalFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfile-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	content := []byte("jobs:\n  job: {}\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	client := NewClient("")
	data, err := client.DownloadWorkflow(tmpfile.Name())
	if err != nil {
		t.Errorf("Expected no error for local file, got: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected file content, got: %s", string(data))
	}
}

func TestDownloadWorkflow_FileURL(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfile-url-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	content := []byte("jobs:\n  job: {}\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	client := NewClient("")
	data, err := client.DownloadWorkflow("file://" + tmpfile.Name())
	if err != nil {
		t.Errorf("Expected no error for file://, got: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected file content, got: %s", string(data))
	}
}

func TestDownloadWorkflow_InvalidPath(t *testing.T) {
	client := NewClient("")
	_, err := client.DownloadWorkflow("/nonexistent/file/path.yml")
	if err == nil {
		t.Errorf("Expected error for invalid file path")
	}
}

func TestDownloadWorkflow_HTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("jobs:\n  job: {}\n")); err != nil {
			t.Errorf("Failed to write HTTP response: %v", err)
		}
	}))
	defer ts.Close()
	client := NewClient("")
	client.httpClient = ts.Client()
	data, err := client.DownloadWorkflow(ts.URL)
	if err != nil {
		t.Errorf("Expected no error for HTTP download, got: %v", err)
	}
	if string(data) != "jobs:\n  job: {}\n" {
		t.Errorf("Expected HTTP content, got: %s", string(data))
	}
}

func TestDownloadWorkflow_HTTPErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("not found")); err != nil {
			t.Errorf("Failed to write HTTP response: %v", err)
		}
	}))
	defer ts.Close()
	client := NewClient("")
	client.httpClient = ts.Client()
	_, err := client.DownloadWorkflow(ts.URL)
	if err == nil {
		t.Errorf("Expected error for HTTP 404")
	}
}
