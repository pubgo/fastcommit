package utils

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDirty(t *testing.T) {
	assert.NoError(t, IsDirty().GetErr())
}

func TestRunGit(t *testing.T) {
	ctx := context.Background()

	t.Run("runs a real git command", func(t *testing.T) {
		err := RunGit(ctx, "--version")
		assert.NoError(t, err)
	})

	t.Run("fails on unknown git subcommand", func(t *testing.T) {
		err := RunGit(ctx, "definitely-not-a-git-subcommand")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "git definitely-not-a-git-subcommand"))
	})
}
