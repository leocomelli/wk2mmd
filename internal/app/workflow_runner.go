package app

import (
	"fmt"
	"log/slog"

	"github.com/leocomelli/wk2mmd/internal/diagram"
	"github.com/leocomelli/wk2mmd/internal/github"
)

// WorkflowRunner encapsulates the logic for analyzing workflows.
type WorkflowRunner struct {
	client github.WorkflowDownloader
}

// NewWorkflowRunner creates a WorkflowRunner for normal use.
func NewWorkflowRunner(token string) *WorkflowRunner {
	return &WorkflowRunner{client: github.NewClient(token)}
}

// NewWorkflowRunnerWithClient creates a WorkflowRunner for testing.
func NewWorkflowRunnerWithClient(client github.WorkflowDownloader) *WorkflowRunner {
	return &WorkflowRunner{client: client}
}

// RunWorkflowAnalysis orchestrates the download, parsing, recursive fetch, and tree/mermaid generation.
func (wr *WorkflowRunner) RunWorkflowAnalysis(workflowURL string, depth int, diagramType string) (string, error) {

	data, err := wr.client.DownloadWorkflow(workflowURL)
	if err != nil {
		return "", fmt.Errorf("failed to download workflow: %w", err)
	}
	slog.Debug("Workflow content", "content", string(data[:min(300, len(data))]))

	wf, err := github.ParseWorkflowYAML(workflowURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Recursively collect all uses and build the tree
	owner, repo, branch := extractRepoInfo(workflowURL)
	slog.Debug("Extracted repo info", "owner", owner, "repo", repo, "branch", branch)

	fetcher := func(uses string) *github.Workflow {
		ar, ok := github.ParseActionRef(uses, owner, repo, branch)
		if !ok {
			return nil
		}
		wf := github.FetchActionWorkflow(wr.client, ar)
		if wf != nil {
			slog.Debug("Fetched reusable workflow", "uses", uses, "jobs", len(wf.Jobs))
		} else {
			slog.Error("Failed to fetch reusable workflow", "uses", uses)
		}
		return wf
	}
	allUses := github.CollectAllUses(wf, fetcher, depth)

	slog.Info("All uses found recursively", "uses", len(allUses))

	tree := github.BuildUsesTree("workflow", wf, fetcher, depth, map[string]bool{})

	// Mermaid diagram generation
	fmt.Println("\nMermaid diagram:")
	switch diagramType {
	case "sequence":
		return diagram.GenerateMermaidSequence(tree), nil
	case "flowchart":
		return diagram.GenerateMermaidFlowchart(tree), nil
	default:
		return "", fmt.Errorf("invalid diagram type: %s", diagramType)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractRepoInfo tries to extract owner, repo, branch from a raw.githubusercontent.com URL.
func extractRepoInfo(url string) (owner, repo, branch string) {
	re := github.ExtractRepoInfoRegex()
	matches := re.FindStringSubmatch(url)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}
