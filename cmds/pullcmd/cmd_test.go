package pullcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildEditorCommand(t *testing.T) {
	t.Run("editor with args", func(t *testing.T) {
		args := buildEditorCommand("code -w", "a.txt")
		require.Equal(t, []string{"code", "-w", "a.txt"}, args)
	})

	t.Run("editor without args", func(t *testing.T) {
		args := buildEditorCommand("vim", "a.txt")
		require.Equal(t, []string{"vim", "a.txt"}, args)
	})
}

func TestSplitRemoteRef(t *testing.T) {
	t.Run("standard origin branch", func(t *testing.T) {
		remote, branch := splitRemoteRef("origin/main")
		require.Equal(t, "origin", remote)
		require.Equal(t, "main", branch)
	})

	t.Run("nested remote branch", func(t *testing.T) {
		remote, branch := splitRemoteRef("upstream/feature/demo")
		require.Equal(t, "upstream", remote)
		require.Equal(t, "feature/demo", branch)
	})
}

func TestValidatePullFlags(t *testing.T) {
	t.Run("no flags is valid", func(t *testing.T) {
		require.NoError(t, validatePullFlags(false, false, false))
	})

	t.Run("rebase alone is valid", func(t *testing.T) {
		require.NoError(t, validatePullFlags(false, false, true))
	})

	t.Run("hard with all is invalid", func(t *testing.T) {
		require.Error(t, validatePullFlags(true, true, false))
	})

	t.Run("rebase with hard is invalid", func(t *testing.T) {
		require.Error(t, validatePullFlags(false, true, true))
	})

	t.Run("rebase with all is invalid", func(t *testing.T) {
		require.Error(t, validatePullFlags(true, false, true))
	})
}

func TestPullExtraArgs(t *testing.T) {
	t.Run("no rebase adds nothing", func(t *testing.T) {
		require.Nil(t, pullExtraArgs(false))
	})

	t.Run("rebase adds --rebase", func(t *testing.T) {
		require.Equal(t, []string{"--rebase"}, pullExtraArgs(true))
	})
}
