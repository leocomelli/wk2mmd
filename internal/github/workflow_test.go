package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWorkflowYAML_NeedsString(t *testing.T) {
	yamlData := []byte(`
jobs:
  build:
    needs: test
    steps:
      - uses: actions/checkout@v2
`)
	wf, err := ParseWorkflowYAML("", yamlData)
	assert.NoError(t, err)
	assert.Contains(t, wf.Jobs, "build")
	assert.Equal(t, NeedsList{"test"}, wf.Jobs["build"].Needs)
	assert.Equal(t, "actions/checkout@v2", wf.Jobs["build"].Steps[0].Uses)
}

func TestParseWorkflowYAML_NeedsArray(t *testing.T) {
	yamlData := []byte(`
jobs:
  deploy:
    needs: [build, test]
    steps:
      - uses: actions/deploy@v1
`)
	wf, err := ParseWorkflowYAML("", yamlData)
	assert.NoError(t, err)
	assert.Contains(t, wf.Jobs, "deploy")
	assert.Equal(t, NeedsList{"build", "test"}, wf.Jobs["deploy"].Needs)
	assert.Equal(t, "actions/deploy@v1", wf.Jobs["deploy"].Steps[0].Uses)
}

func TestParseWorkflowYAML_InvalidYAML(t *testing.T) {
	yamlData := []byte(`invalid: [unclosed`)
	_, err := ParseWorkflowYAML("", yamlData)
	assert.Error(t, err)
}

func TestParseWorkflowYAML_InvalidNeeds(t *testing.T) {
	yamlData := []byte(`
jobs:
  build:
    needs: {foo: bar}
    steps:
      - uses: actions/checkout@v2
`)
	_, err := ParseWorkflowYAML("", yamlData)
	assert.Error(t, err)
}

func TestParseWorkflowYAML_JobLevelUses(t *testing.T) {
	yamlData := []byte(`
jobs:
  call_another_workflow:
    uses: owner/repo/.github/workflows/workflow.yml@main
`)
	wf, err := ParseWorkflowYAML("", yamlData)
	assert.NoError(t, err)
	assert.Contains(t, wf.Jobs, "call_another_workflow")
	assert.Equal(t, "owner/repo/.github/workflows/workflow.yml@main", wf.Jobs["call_another_workflow"].Uses)
}

func TestParseActionRef_Remote(t *testing.T) {
	ar, ok := ParseActionRef("octocat/myaction/path@v1", "", "", "")
	assert.True(t, ok)
	assert.Equal(t, "remote", ar.Type)

	assert.Equal(t, "", ar.Owner)
	assert.Equal(t, "", ar.Repo)
	assert.Equal(t, "octocat/myaction/path@v1", ar.Path)
	assert.Equal(t, "", ar.Ref)
}

func TestParseActionRef_Marketplace(t *testing.T) {
	ar, ok := ParseActionRef("actions/checkout@v2", "", "", "")
	assert.True(t, ok)

	assert.Equal(t, "remote", ar.Type)
}

func TestParseActionRef_Unrecognized(t *testing.T) {
	_, ok := ParseActionRef("echo hello", "", "", "")
	assert.True(t, ok)
}

type mockClient struct {
	DownloadWorkflowFunc func(url string) ([]byte, error)
}

func (m *mockClient) DownloadWorkflow(url string) ([]byte, error) {
	return m.DownloadWorkflowFunc(url)
}

func TestFetchActionWorkflow_MarketplaceOrUnknown(t *testing.T) {
	client := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return nil, assert.AnError
		},
	}
	ar := ActionRef{Type: "marketplace"}
	assert.Nil(t, FetchActionWorkflow(client, ar))
	ar.Type = "unknown"
	assert.Nil(t, FetchActionWorkflow(client, ar))
}

func TestFetchActionWorkflow_ErrorCases(t *testing.T) {
	client := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return nil, assert.AnError
		},
	}
	ar := ActionRef{Type: "local", Path: "some/path"}
	wf := FetchActionWorkflow(client, ar)
	if wf != nil {
		t.Errorf("Expected nil when download fails")
	}

	client2 := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return []byte("invalid: [unclosed"), nil
		},
	}
	wf2 := FetchActionWorkflow(client2, ar)
	if wf2 != nil {
		t.Errorf("Expected nil when parsing fails")
	}

	client3 := &mockClient{
		DownloadWorkflowFunc: func(url string) ([]byte, error) {
			return []byte("name: Empty"), nil
		},
	}
	wf3 := FetchActionWorkflow(client3, ar)
	if wf3 != nil {
		t.Errorf("Expected nil when workflow has no jobs")
	}
}

func TestBuildUsesTree_Simple(t *testing.T) {
	wf := &Workflow{
		Jobs: map[string]Job{
			"a": {},
			"b": {},
		},
	}
	tree := BuildUsesTree("root", wf, nil, 2, map[string]bool{})
	if tree == nil {
		t.Errorf("Expected non-nil tree")
		return
	}
	if len(tree.Children) != 2 {
		t.Errorf("Expected 2 children, got %v", len(tree.Children))
	}
	names := []string{tree.Children[0].Name, tree.Children[1].Name}
	if !contains(names, "a") || !contains(names, "b") {
		t.Errorf("Expected children 'a' and 'b', got %v", names)
	}
}

func TestBuildUsesTree_WithReusable(t *testing.T) {
	// Simula um workflow com um job que usa outro workflow (reusable)
	wf := &Workflow{
		Jobs: map[string]Job{
			"main": {Uses: "reusable.yml@main"},
		},
	}
	fetcher := func(uses string) *Workflow {
		if uses == "reusable.yml@main" {
			return &Workflow{Jobs: map[string]Job{"a": {}, "b": {}}}
		}
		return nil
	}
	tree := BuildUsesTree("root", wf, fetcher, 2, map[string]bool{})
	if tree == nil || len(tree.Children) != 1 {
		t.Errorf("Expected 1 child, got %v", len(tree.Children))
	}
	child := tree.Children[0]
	if child.Name != "main" || len(child.Children) != 2 {
		t.Errorf("Expected child 'main' with 2 subjobs, got %v with %d", child.Name, len(child.Children))
	}
}

func TestBuildUsesTree_DepthLimit(t *testing.T) {
	wf := &Workflow{
		Jobs: map[string]Job{
			"main": {Uses: "reusable.yml@main"},
		},
	}
	fetcher := func(uses string) *Workflow {
		return &Workflow{Jobs: map[string]Job{"a": {}, "b": {}}}
	}
	tree := BuildUsesTree("root", wf, fetcher, 1, map[string]bool{})
	if tree == nil || len(tree.Children) != 1 {
		t.Errorf("Expected 1 child at depth 1, got %v", len(tree.Children))
	}
	if len(tree.Children[0].Children) != 0 {
		t.Errorf("Expected no grandchildren at depth 1, got %d", len(tree.Children[0].Children))
	}
}

func TestBuildUsesTree_Cycle(t *testing.T) {
	wf := &Workflow{
		Jobs: map[string]Job{
			"main": {Uses: "reusable.yml@main"},
		},
	}
	fetcher := func(uses string) *Workflow {
		if uses == "reusable.yml@main" {
			return wf
		}
		return nil
	}
	tree := BuildUsesTree("root", wf, fetcher, 5, map[string]bool{})
	if tree == nil || len(tree.Children) != 1 {
		t.Errorf("Expected 1 child, got %v", len(tree.Children))
	}
}

func TestBuildUsesTree_EmptyWorkflow(t *testing.T) {
	wf := &Workflow{Jobs: map[string]Job{}}
	tree := BuildUsesTree("root", wf, nil, 2, map[string]bool{})
	if tree == nil {
		t.Errorf("Expected non-nil tree for empty workflow")
		return
	}
	if tree != nil && len(tree.Children) != 0 {
		t.Errorf("Expected no children for empty workflow, got %d", len(tree.Children))
	}
}

func TestCollectAllUses_JobLevel(t *testing.T) {
	wf := &Workflow{
		Jobs: map[string]Job{
			"job1": {Uses: "owner/repo/.github/workflows/workflow.yml@main"},
			"job2": {Uses: "actions/checkout@v4"},
		},
	}
	uses := CollectAllUses(wf, nil, 2)
	if len(uses) != 2 {
		t.Errorf("Expected 2 uses, got %v", uses)
	}
	if !contains(uses, "owner/repo/.github/workflows/workflow.yml@main") || !contains(uses, "actions/checkout@v4") {
		t.Errorf("Expected both uses, got %v", uses)
	}
}

func TestCollectAllUses_DepthLimit(t *testing.T) {
	wf := &Workflow{
		Jobs: map[string]Job{
			"main": {Uses: "reusable.yml@main"},
		},
	}
	fetcher := func(uses string) *Workflow {
		return &Workflow{Jobs: map[string]Job{"a": {Uses: "other.yml@main"}}}
	}
	uses := CollectAllUses(wf, fetcher, 1)
	if len(uses) != 1 || uses[0] != "reusable.yml@main" {
		t.Errorf("Expected only the first level use, got %v", uses)
	}
	uses2 := CollectAllUses(wf, fetcher, 2)
	if len(uses2) != 2 {
		t.Errorf("Expected two uses with depth 2, got %v", uses2)
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestExtractRepoInfoRegex(t *testing.T) {
	re := ExtractRepoInfoRegex()
	url := "https://raw.githubusercontent.com/owner/repo/branch/path/to/file.yml"
	matches := re.FindStringSubmatch(url)
	if len(matches) != 4 {
		t.Errorf("Expected 4 matches, got %v", matches)
	}
	if matches[1] != "owner" || matches[2] != "repo" || matches[3] != "branch" {
		t.Errorf("Unexpected extraction: %v", matches)
	}
}
