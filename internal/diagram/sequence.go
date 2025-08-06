package diagram

import (
	"github.com/TyphonHill/go-mermaid/diagrams/sequence"
	"github.com/leocomelli/wk2mmd/internal/github"
)

// GenerateMermaidSequence generates a Mermaid sequence from a UsesNode tree using go-mermaid.
func GenerateMermaidSequence(root *github.UsesNode) string {
	diagram := sequence.NewDiagram()
	nodeMap := make(map[string]*sequence.Actor)
	buildSequenceActors(diagram, root, nodeMap)
	addSequenceMessages(diagram, root, nodeMap)
	return diagram.String()
}

func buildSequenceActors(diagram *sequence.Diagram, node *github.UsesNode, nodeMap map[string]*sequence.Actor) {
	if node == nil {
		return
	}
	if _, exists := nodeMap[node.UniqueID]; !exists {
		nodeMap[node.UniqueID] = diagram.AddActor(node.UniqueID, node.Name, sequence.ActorParticipant)
	}
	for _, child := range node.Children {
		buildSequenceActors(diagram, child, nodeMap)
	}
}

func addSequenceMessages(diagram *sequence.Diagram, node *github.UsesNode, nodeMap map[string]*sequence.Actor) {
	if node == nil {
		return
	}
	from := nodeMap[node.UniqueID]
	for _, child := range node.Children {
		to := nodeMap[child.UniqueID]
		if from != nil && to != nil {
			diagram.AddMessage(from, to, sequence.MessageSolid, "uses")
		}
		addSequenceMessages(diagram, child, nodeMap)
	}
}
