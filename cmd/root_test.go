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
