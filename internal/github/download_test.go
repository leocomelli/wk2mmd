package github

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadWorkflow(t *testing.T) {
	// Create a test server that returns a fake workflow
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("name: Test Workflow\n"))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient("")
	// Override httpClient to use the test server
	client.httpClient = ts.Client()

	data, err := client.DownloadWorkflow(ts.URL)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Test Workflow")
}

func TestDownloadWorkflow_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("not found"))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient("")
	client.httpClient = ts.Client()

	_, err := client.DownloadWorkflow(ts.URL)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "unexpected status code")
	}
}
