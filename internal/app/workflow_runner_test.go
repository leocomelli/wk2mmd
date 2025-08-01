package app

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	DownloadWorkflowFunc func(url string) ([]byte, error)
}

func (m *mockClient) DownloadWorkflow(url string) ([]byte, error) {
	return m.DownloadWorkflowFunc(url)
}

func TestRunWorkflowAnalysis_Success(t *testing.T) {
	client := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return []byte(`jobs: { job: { steps: [ { uses: "x" } ] } }`), nil
		},
	}
	runner := NewWorkflowRunnerWithClient(client)
	err := runner.RunWorkflowAnalysis("https://raw.githubusercontent.com/owner/repo/branch/file.yml", 2, "flowchart")
	assert.NoError(t, err)
}

func TestRunWorkflowAnalysis_DownloadFail(t *testing.T) {
	client := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return nil, errors.New("fail download")
		},
	}
	runner := NewWorkflowRunnerWithClient(client)
	err := runner.RunWorkflowAnalysis("https://raw.githubusercontent.com/owner/repo/branch/file.yml", 2, "flowchart")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download workflow")
}

func TestRunWorkflowAnalysis_ParseFail(t *testing.T) {
	client := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return []byte("invalid: [unclosed"), nil
		},
	}
	runner := NewWorkflowRunnerWithClient(client)
	err := runner.RunWorkflowAnalysis("https://raw.githubusercontent.com/owner/repo/branch/file.yml", 2, "flowchart")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse workflow YAML")
}
