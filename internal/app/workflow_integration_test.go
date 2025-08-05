//go:build integration
// +build integration

package app

import (
	"sort"
	"strings"
	"testing"
)

func TestRunWorkflowAnalysis_Integration(t *testing.T) {

	runner := NewWorkflowRunner("")
	output, err := runner.RunWorkflowAnalysis("https://github.com/leocomelli/wk2mmd/blob/main/.github/test-workflows/reusable-wf1.yml", 10, "flowchart")
	if err != nil {
		t.Fatalf("Failed to run workflow analysis: %v", err)
	}

	expected := `
---
title: Workflow Graph
config:
    theme: default
    maxTextSize: 50000
    maxEdges: 500
    fontSize: 16
---
flowchart TB
    0@{ shape: rect, label: "workflow"}
    1@{ shape: rect, label: "job_a"}
    2@{ shape: rect, label: "job_b"}
    3@{ shape: rect, label: "call_workflow_2"}
    4@{ shape: rect, label: "prepare"}
    5@{ shape: rect, label: "build"}
    6@{ shape: rect, label: "call_workflow_3"}
    7@{ shape: rect, label: "setup"}
    8@{ shape: rect, label: "test"}
    9@{ shape: rect, label: "call_workflow_4"}
    10@{ shape: rect, label: "init"}
    11@{ shape: rect, label: "verify"}
    12@{ shape: rect, label: "call_workflow_5"}
    13@{ shape: rect, label: "start"}
    14@{ shape: rect, label: "finalize"}
    0 --> 1
    0 --> 2
    0 --> 3
    3 --> 4
    3 --> 5
    3 --> 6
    6 --> 7
    6 --> 8
    6 --> 9
    9 --> 10
    9 --> 11
    9 --> 12
    12 --> 13
    12 --> 14
`
	if !compareMermaidIgnoringOrder(output, expected) {
		t.Errorf("Flowchart output does not match expected (order-agnostic).\nGot:\n%s\nExpected:\n%s", output, expected)
	}
}

func compareMermaidIgnoringOrder(a, b string) bool {
	alines := filterLines(a)
	blines := filterLines(b)
	sort.Strings(alines)
	sort.Strings(blines)
	if len(alines) != len(blines) {
		return false
	}
	for i := range alines {
		if alines[i] != blines[i] {
			return false
		}
	}
	return true
}

func filterLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			lines = append(lines, trimmed)
		}
	}
	return lines
}
