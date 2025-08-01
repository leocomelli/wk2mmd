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
	wf, err := ParseWorkflowYAML(yamlData)
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
	wf, err := ParseWorkflowYAML(yamlData)
	assert.NoError(t, err)
	assert.Contains(t, wf.Jobs, "deploy")
	assert.Equal(t, NeedsList{"build", "test"}, wf.Jobs["deploy"].Needs)
	assert.Equal(t, "actions/deploy@v1", wf.Jobs["deploy"].Steps[0].Uses)
}

func TestParseWorkflowYAML_StepsNoUses(t *testing.T) {
	yamlData := []byte(`
jobs:
  test:
    steps:
      - name: Run tests
        run: go test ./...
`)
	wf, err := ParseWorkflowYAML(yamlData)
	assert.NoError(t, err)
	assert.Contains(t, wf.Jobs, "test")
	assert.Equal(t, "Run tests", wf.Jobs["test"].Steps[0].Name)
	assert.Equal(t, "go test ./...", wf.Jobs["test"].Steps[0].Run)
	assert.Equal(t, "", wf.Jobs["test"].Steps[0].Uses)
}

func TestParseWorkflowYAML_InvalidYAML(t *testing.T) {
	yamlData := []byte(`invalid: [unclosed`)
	_, err := ParseWorkflowYAML(yamlData)
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
	_, err := ParseWorkflowYAML(yamlData)
	assert.Error(t, err)
}

func TestCollectAllUses_Recursive(t *testing.T) {
	// Simulate three levels: workflow -> action1 -> action2
	mainWf := &Workflow{
		Jobs: map[string]Job{
			"build": {
				Steps: []Step{{Uses: "./.github/actions/action1"}},
			},
		},
	}
	action1 := &Workflow{
		Jobs: map[string]Job{
			"action-job": {
				Steps: []Step{{Uses: "./.github/actions/action2"}},
			},
		},
	}
	action2 := &Workflow{
		Jobs: map[string]Job{
			"action-job": {
				Steps: []Step{{Uses: "actions/checkout@v2"}},
			},
		},
	}
	fakeFetcher := func(uses string) *Workflow {
		switch uses {
		case "./.github/actions/action1":
			return action1
		case "./.github/actions/action2":
			return action2
		default:
			return nil
		}
	}
	allUses := CollectAllUses(mainWf, fakeFetcher, 3)
	assert.Contains(t, allUses, "./.github/actions/action1")
	assert.Contains(t, allUses, "./.github/actions/action2")
	assert.Contains(t, allUses, "actions/checkout@v2")
	assert.Equal(t, 3, len(allUses))
}

func TestParseActionRef_Local(t *testing.T) {
	ar, ok := ParseActionRef("./.github/actions/foo", "me", "repo", "main")
	assert.True(t, ok)
	assert.Equal(t, "local", ar.Type)
	assert.Equal(t, "me", ar.Owner)
	assert.Equal(t, "repo", ar.Repo)
	assert.Equal(t, "main", ar.Ref)
	assert.Equal(t, "actions/foo", ar.Path)
}

func TestParseActionRef_Remote(t *testing.T) {
	ar, ok := ParseActionRef("octocat/myaction/path@v1", "", "", "")
	assert.True(t, ok)
	assert.Equal(t, "remote", ar.Type)
	assert.Equal(t, "octocat", ar.Owner)
	assert.Equal(t, "myaction", ar.Repo)
	assert.Equal(t, "path", ar.Path)
	assert.Equal(t, "v1", ar.Ref)
}

func TestParseActionRef_Marketplace(t *testing.T) {
	ar, ok := ParseActionRef("actions/checkout@v2", "", "", "")
	assert.True(t, ok)
	assert.Equal(t, "marketplace", ar.Type)
}

func TestParseActionRef_Unrecognized(t *testing.T) {
	_, ok := ParseActionRef("echo hello", "", "", "")
	assert.False(t, ok)
}
