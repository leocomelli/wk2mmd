package diagram

import (
	"github.com/TyphonHill/go-mermaid/diagrams/flowchart"
	"github.com/leocomelli/wk2mmd/internal/github"
)

// GenerateMermaidFlowchart generates a Mermaid flowchart (TD) from a UsesNode tree using go-mermaid.
func GenerateMermaidFlowchart(root *github.UsesNode) string {
	fc := flowchart.NewFlowchart()
	fc.Title = "Workflow Graph"

	nodeMap := make(map[string]*flowchart.Node)
	buildFlowchartNodes(fc, root, nodeMap)
	addFlowchartLinks(fc, root, nodeMap)

	return fc.String()
}

// buildFlowchartNodes recursively adds nodes to the flowchart.
func buildFlowchartNodes(fc *flowchart.Flowchart, node *github.UsesNode, nodeMap map[string]*flowchart.Node) {
	if node == nil {
		return
	}
	if _, exists := nodeMap[node.Name]; !exists {
		nodeMap[node.Name] = fc.AddNode(node.Name)
	}
	for _, child := range node.Children {
		buildFlowchartNodes(fc, child, nodeMap)
	}
}

// addFlowchartLinks recursively adds links between nodes to the flowchart.
func addFlowchartLinks(fc *flowchart.Flowchart, node *github.UsesNode, nodeMap map[string]*flowchart.Node) {
	if node == nil {
		return
	}
	from := nodeMap[node.Name]
	for _, child := range node.Children {
		to := nodeMap[child.Name]
		if from != nil && to != nil {
			fc.AddLink(from, to)
		}
		addFlowchartLinks(fc, child, nodeMap)
	}
}
