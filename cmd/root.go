package cmd

import (
	"fmt"
	"os"

	"github.com/leocomelli/wk2mmd/internal/app"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowURL := args[0]
		runner := app.NewWorkflowRunner(token)
		output, err := runner.RunWorkflowAnalysis(workflowURL, depth, diagramType)
		if err != nil {
			return err
		}

		fmt.Println(output)

		return nil
	},
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
