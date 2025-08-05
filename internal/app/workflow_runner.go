package app

import (
	"fmt"

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
func (wr *WorkflowRunner) RunWorkflowAnalysis(workflowURL string, depth int, diagramType string) error {
	fmt.Println("Diagram type:", diagramType)
	fmt.Println("Depth:", depth)
	fmt.Println("Workflow URL:", workflowURL)

	data, err := wr.client.DownloadWorkflow(workflowURL)
	if err != nil {
		return fmt.Errorf("failed to download workflow: %w", err)
	}
	fmt.Printf("Workflow content (first 300 bytes):\n%s\n", string(data[:min(300, len(data))]))

	wf, err := github.ParseWorkflowYAML(data)
	if err != nil {
		return fmt.Errorf("failed to parse workflow YAML: %w", err)
	}
	fmt.Println("Jobs found:")
	for jobName, job := range wf.Jobs {
		fmt.Printf("- Job: %s\n", jobName)
		if len(job.Needs) > 0 {
			fmt.Printf("  Needs: %v\n", job.Needs)
		}
		for i, step := range job.Steps {
			if step.Uses != "" {
				fmt.Printf("  Step %d uses: %s\n", i+1, step.Uses)
			}
		}
	}

	// Recursively collect all uses and build the tree
	owner, repo, branch := extractRepoInfo(workflowURL)
	fetcher := func(uses string) *github.Workflow {
		ar, ok := github.ParseActionRef(uses, owner, repo, branch)
		if !ok {
			return nil
		}
		wf := github.FetchActionWorkflow(wr.client, ar)
		if wf != nil {
			fmt.Printf("[DEBUG] Fetched reusable workflow: %s (jobs: %d)\n", uses, len(wf.Jobs))
		} else {
			fmt.Printf("[DEBUG] Failed to fetch reusable workflow: %s\n", uses)
		}
		return wf
	}
	allUses := github.CollectAllUses(wf, fetcher, depth)
	fmt.Println("All uses found recursively:")
	for _, u := range allUses {
		fmt.Println("-", u)
	}

	tree := github.BuildUsesTree("workflow", wf, fetcher, depth, map[string]bool{})
	fmt.Println("Uses tree:")
	printUsesTree(tree, 0)

	// Mermaid diagram generation
	fmt.Println("\nMermaid diagram:")
	if diagramType == "sequence" {
		fmt.Println(diagram.GenerateMermaidSequence(tree))
	} else {
		fmt.Println(diagram.GenerateMermaidFlowchart(tree))
	}
	return nil
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

func printUsesTree(node *github.UsesNode, indent int) {
	if node == nil {
		return
	}
	fmt.Printf("%s- %s\n", spaces(indent), node.Name)
	for _, child := range node.Children {
		printUsesTree(child, indent+2)
	}
}

func spaces(n int) string {
	return fmt.Sprintf("%*s", n, "")
}
