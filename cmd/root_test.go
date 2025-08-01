package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 2, min(2, 2))
	assert.Equal(t, -1, min(-1, 2))
	assert.Equal(t, 0, min(0, 5))
}

func TestToRawGitHubURL(t *testing.T) {
	// Should convert /blob/ URL
	input := "https://github.com/owner/repo/blob/branch/path/to/file.yml"
	expected := "https://raw.githubusercontent.com/owner/repo/branch/path/to/file.yml"
	assert.Equal(t, expected, toRawGitHubURL(input))

	// Should not convert already raw URL
	raw := "https://raw.githubusercontent.com/owner/repo/branch/path/to/file.yml"
	assert.Equal(t, raw, toRawGitHubURL(raw))

	// Should not convert unrelated URL
	other := "https://example.com/some/file.yml"
	assert.Equal(t, other, toRawGitHubURL(other))
}

func TestExecute_NoArgs(t *testing.T) {
	// Save and restore os.Args and os.Exit
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"wk2mmd"}

	cmd := rootCmd
	cmd.SetArgs([]string{})
	// Cobra will call os.Exit(1), so we recover from panic
	defer func() {
		_ = recover()
	}()
	err := cmd.Execute()
	assert.Error(t, err)
}
