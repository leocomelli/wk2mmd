//go:build integration
// +build integration

package app

import (
	"fmt"
	"testing"
)

func TestRunWorkflowAnalysis_Integration(t *testing.T) {

	runner := NewWorkflowRunner("")

	workflowURL := "https://github.com/leocomelli/wk2mmd/blob/main/.github/test-workflows/reusable-wf1.yml"
	diagramType := "flowchart"
	depth := 10

	output, err := runner.RunWorkflowAnalysis(workflowURL, depth, diagramType)
	if err != nil {
		t.Fatalf("Failed to run workflow analysis: %v", err)
	}

	fmt.Println("--------------------------------")
	fmt.Println(output)
}
