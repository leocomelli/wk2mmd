package github

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job represents a job in a GitHub Actions workflow.
type Job struct {
	Needs NeedsList `yaml:"needs"`
	Steps []Step    `yaml:"steps"`
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
	if strings.HasPrefix(uses, "./") || strings.HasPrefix(uses, ".github/") {
		// Local action
		ar.Type = "local"
		ar.Owner = repoOwner
		ar.Repo = repoName
		ar.Ref = branch
		ar.Path = strings.TrimPrefix(uses, "./")
		ar.Path = strings.TrimPrefix(ar.Path, ".github/")
		return ar, true
	}
	// Marketplace action: actions/checkout@v2
	if strings.HasPrefix(uses, "actions/") && strings.Contains(uses, "@") {
		ar.Type = "marketplace"
		return ar, true
	}
	// Remote action: owner/repo/path@ref
	re := regexp.MustCompile(`^([^/]+)/([^/@]+)(/[^@]*)?@(.+)$`)
	matches := re.FindStringSubmatch(uses)
	if len(matches) == 5 {
		ar.Type = "remote"
		ar.Owner = matches[1]
		ar.Repo = matches[2]
		ar.Path = strings.TrimPrefix(matches[3], "/")
		ar.Ref = matches[4]
		return ar, true
	}
	return ActionRef{}, false
}

// FetchActionWorkflow tries to download and parse the action.yml or action.yaml for a given ActionRef.
// Returns the parsed Workflow or nil if not found or not a composite action.
func FetchActionWorkflow(client *Client, ar ActionRef) *Workflow {
	var urls []string
	switch ar.Type {
	case "local":
		// Local action: https://raw.githubusercontent.com/{owner}/{repo}/{ref}/{path}/action.yml
		base := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", ar.Owner, ar.Repo, ar.Ref, strings.TrimSuffix(ar.Path, "/"))
		urls = []string{base + "/action.yml", base + "/action.yaml"}
	case "remote":
		// Remote action: https://raw.githubusercontent.com/{owner}/{repo}/{ref}/{path}/action.yml
		base := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", ar.Owner, ar.Repo, ar.Ref, strings.TrimSuffix(ar.Path, "/"))
		urls = []string{base + "/action.yml", base + "/action.yaml"}
	default:
		return nil
	}
	for _, url := range urls {
		data, err := client.DownloadWorkflow(url)
		if err != nil {
			continue
		}
		wf, err := ParseWorkflowYAML(data)
		if err == nil && wf != nil && len(wf.Jobs) > 0 {
			return wf
		}
	}
	return nil
}

// CollectAllUses recursively collects all 'uses' from a workflow and its referenced actions, up to a given depth.
// The fetcher function should return a Workflow for a given uses string (e.g., a custom action path).
func CollectAllUses(wf *Workflow, fetcher func(string) *Workflow, depth int) []string {
	if depth == 0 || wf == nil {
		return nil
	}
	var uses []string
	for _, job := range wf.Jobs {
		for _, step := range job.Steps {
			if step.Uses != "" {
				uses = append(uses, step.Uses)
				if fetcher != nil {
					childWf := fetcher(step.Uses)
					if childWf != nil {
						uses = append(uses, CollectAllUses(childWf, fetcher, depth-1)...)
					}
				}
			}
		}
	}
	return uses
}

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
