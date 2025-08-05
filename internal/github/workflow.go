package github

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkflowDownloader defines the interface for downloading workflow files.
type WorkflowDownloader interface {
	DownloadWorkflow(url string) ([]byte, error)
}

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job represents a job in a GitHub Actions workflow.
type Job struct {
	Needs NeedsList `yaml:"needs"`
	Steps []Step    `yaml:"steps"`
	Uses  string    `yaml:"uses"`
}

// Step represents a step in a job.
type Step struct {
	Uses string `yaml:"uses"`
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// NeedsList handles both string and []string for the 'needs' field.
type NeedsList []string

// ActionRef represents a parsed 'uses' reference in a workflow step.
type ActionRef struct {
	Type  string // "local", "remote", or "marketplace"
	Owner string
	Repo  string
	Ref   string
	Path  string
	Raw   string // original uses string
}

// UsesNode represents a node in the uses dependency tree.
type UsesNode struct {
	Name     string
	Children []*UsesNode
}

// ParseWorkflowYAML parses the workflow YAML into a Workflow struct.
func ParseWorkflowYAML(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}
	return &wf, nil
}

// ParseActionRef parses a 'uses' string and returns an ActionRef.
// repoOwner, repoName, branch are used for resolving local actions.
// Returns (ActionRef, true) if recognized, or (zero, false) if not.
func ParseActionRef(uses, repoOwner, repoName, branch string) (ActionRef, bool) {
	ar := ActionRef{Raw: uses}

	// local action
	if strings.HasPrefix(uses, "./") || strings.HasPrefix(uses, ".github/") {
		ar.Type = "local"
		ar.Path = uses
		return ar, true
	}

	// remote action
	ar.Type = "remote"
	ar.Owner = repoOwner
	ar.Repo = repoName
	ar.Ref = branch
	ar.Path = strings.TrimPrefix(uses, fmt.Sprintf("%s/%s/", repoOwner, repoName))
	return ar, true
}

// FetchActionWorkflow tries to download and parse the action.yml or action.yaml for a given ActionRef.
// Returns the parsed Workflow or nil if not found or not a composite action.
func FetchActionWorkflow(client WorkflowDownloader, ar ActionRef) *Workflow {
	var urls []string
	switch ar.Type {
	case "local":
		urls = []string{ar.Path}
	case "remote":
		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/%s/%s", ar.Owner, ar.Repo, ar.Ref, strings.TrimSuffix(ar.Path, "/"))
		urls = []string{url}
	default:
		return nil
	}
	for _, url := range urls {
		data, err := client.DownloadWorkflow(url)
		if err != nil {
			slog.Error("Failed to download workflow", "url", url, "error", err)
			continue
		}
		wf, err := ParseWorkflowYAML(data)
		if err == nil && wf != nil && len(wf.Jobs) > 0 {
			return wf
		}
	}
	return nil
}

// BuildUsesTree builds a hierarchical tree of uses dependencies starting from the given workflow.
func BuildUsesTree(name string, wf *Workflow, fetcher func(string) *Workflow, depth int, visited map[string]bool) *UsesNode {
	if depth == 0 || wf == nil || visited[name] {
		return nil
	}
	visited[name] = true

	node := &UsesNode{Name: name}
	for jobName, job := range wf.Jobs {
		if job.Uses != "" {
			child := &UsesNode{Name: jobName}
			if fetcher != nil && depth > 1 {
				childWf := fetcher(job.Uses)
				if childWf != nil {
					for subJobName, subJob := range childWf.Jobs {
						if subJob.Uses != "" && fetcher != nil && depth > 2 {
							subChildWf := fetcher(subJob.Uses)
							subChild := &UsesNode{Name: subJobName}
							if subChildWf != nil {
								subtree := BuildUsesTree(subJobName, subChildWf, fetcher, depth-2, visited)
								if subtree != nil {
									subChild.Children = subtree.Children
								}
							}
							child.Children = append(child.Children, subChild)
						} else {
							child.Children = append(child.Children, &UsesNode{Name: subJobName})
						}
					}
				}
			}
			node.Children = append(node.Children, child)
			continue
		}
		// If not a reusable, just add the job and its steps
		jobNode := &UsesNode{Name: jobName}
		for _, step := range job.Steps {
			if step.Uses != "" {
				stepNode := &UsesNode{Name: step.Uses}
				if fetcher != nil && depth > 1 {
					childWf := fetcher(step.Uses)
					if childWf != nil {
						subtree := BuildUsesTree(step.Uses, childWf, fetcher, depth-1, visited)
						if subtree != nil {
							stepNode.Children = subtree.Children
						}
					}
				}
				jobNode.Children = append(jobNode.Children, stepNode)
			}
		}
		node.Children = append(node.Children, jobNode)
	}
	return node
}

// CollectAllUses recursively collects all 'uses' from a workflow and its referenced actions, up to a given depth.
func CollectAllUses(wf *Workflow, fetcher func(string) *Workflow, depth int) []string {
	if depth == 0 || wf == nil {
		return nil
	}
	var uses []string
	for _, job := range wf.Jobs {
		// Job-level uses
		if job.Uses != "" {
			uses = append(uses, job.Uses)
			if fetcher != nil {
				childWf := fetcher(job.Uses)
				if childWf != nil {
					uses = append(uses, CollectAllUses(childWf, fetcher, depth-1)...)
				}
			}
		}
	}
	return uses
}

// UnmarshalYAML custom unmarshal for NeedsList to support string or []string.
func (n *NeedsList) UnmarshalYAML(value *yaml.Node) error {
	var single string
	if err := value.Decode(&single); err == nil {
		*n = NeedsList{single}
		return nil
	}
	var multi []string
	if err := value.Decode(&multi); err == nil {
		*n = NeedsList(multi)
		return nil
	}
	return fmt.Errorf("invalid needs field: %v", value.Value)
}

// ExtractRepoInfoRegex returns the regex to extract owner, repo, branch from a raw.githubusercontent.com or github.com/blob URL.
func ExtractRepoInfoRegex() *regexp.Regexp {
	// Supports:
	// https://raw.githubusercontent.com/owner/repo/branch/path/to/file.yml
	// https://github.com/owner/repo/blob/branch/path/to/file.yml
	return regexp.MustCompile(`https://(?:raw\.githubusercontent\.com|github\.com)/([^/]+)/([^/]+)/(?:blob/)?([^/]+)/`)
}
