package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/leocomelli/wk2mmd/internal/github"
	"github.com/spf13/cobra"
)

var (
	diagramType string
	depth       int
	token       string
)

// toRawGitHubURL converts a github.com URL with /blob/ to a raw.githubusercontent.com URL.
func toRawGitHubURL(url string) string {
	if strings.HasPrefix(url, "https://github.com/") && strings.Contains(url, "/blob/") {
		re := regexp.MustCompile(`https://github.com/([^/]+)/([^/]+)/blob/([^/]+)/(.*)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) == 5 {
			owner, repo, branch, path := matches[1], matches[2], matches[3], matches[4]
			return "https://raw.githubusercontent.com/" + owner + "/" + repo + "/" + branch + "/" + path
		}
	}
	return url
}

// extractRepoInfo tries to extract owner, repo, branch from a raw.githubusercontent.com URL.
func extractRepoInfo(url string) (owner, repo, branch string) {
	re := regexp.MustCompile(`https://raw.githubusercontent.com/([^/]+)/([^/]+)/([^/]+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}

var rootCmd = &cobra.Command{
	Use:   "wk2mmd <workflow-url>",
	Short: "Generate a Mermaid diagram from a GitHub Actions workflow file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowURL := toRawGitHubURL(args[0])
		fmt.Println("Diagram type:", diagramType)
		fmt.Println("Depth:", depth)
		fmt.Println("Token provided:", token != "")
		fmt.Println("Workflow URL:", workflowURL)

		client := github.NewClient(token)
		data, err := client.DownloadWorkflow(workflowURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download workflow: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Workflow content (first 300 bytes):\n%s\n", string(data[:min(300, len(data))]))

		wf, err := github.ParseWorkflowYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse workflow YAML: %v\n", err)
			os.Exit(1)
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

		// Recursively collect all uses
		owner, repo, branch := extractRepoInfo(workflowURL)
		fetcher := func(uses string) *github.Workflow {
			ar, ok := github.ParseActionRef(uses, owner, repo, branch)
			if !ok {
				return nil
			}
			return github.FetchActionWorkflow(client, ar)
		}
		allUses := github.CollectAllUses(wf, fetcher, depth)
		fmt.Println("All uses found recursively:")
		for _, u := range allUses {
			fmt.Println("-", u)
		}
		// Next: generate Mermaid diagram
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Execute runs the root command.
func Execute() {
	rootCmd.Flags().StringVarP(&diagramType, "diagram-type", "t", "flowchart", "Mermaid diagram type: flowchart or sequence")
	rootCmd.Flags().IntVarP(&depth, "depth", "d", 2, "Maximum depth for recursive 'uses' analysis")
	rootCmd.Flags().StringVarP(&token, "token", "k", "", "GitHub token for accessing private repositories")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
