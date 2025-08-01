package cmd

import (
	"fmt"
	"os"

	"github.com/leocomelli/wk2mmd/internal/github"
	"github.com/spf13/cobra"
)

var (
	diagramType string
	depth       int
	token       string
)

var rootCmd = &cobra.Command{
	Use:   "wk2mmd <workflow-url>",
	Short: "Generate a Mermaid diagram from a GitHub Actions workflow file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowURL := args[0]
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
		// Next: parse the YAML
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
