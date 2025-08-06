package diagram

import (
	"strings"
	"testing"

	"github.com/leocomelli/wk2mmd/internal/github"
)

func TestGenerateMermaidFlowchart_Simple(t *testing.T) {
	root := &github.UsesNode{
		Name: "root",
		Children: []*github.UsesNode{
			{Name: "a", UniqueID: "root/a"},
			{Name: "b", UniqueID: "root/b"},
		},
		UniqueID: "root",
	}

	result := GenerateMermaidFlowchart(root)
	if !strings.Contains(result, "flowchart") {
		t.Errorf("Expected output to contain 'flowchart', got: %s", result)
	}
	if !strings.Contains(result, "root") || !strings.Contains(result, "a") || !strings.Contains(result, "b") {
		t.Errorf("Expected output to contain all node names, got: %s", result)
	}
}
