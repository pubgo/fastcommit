package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitCommitRejectsEmptyMessage(t *testing.T) {
	err := GitCommit(context.Background(), "   ")
	require.Error(t, err)
}

func TestGitCommitCreatesCommit(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.txt"), []byte("hello\n"), 0o644))
	run("add", "a.txt")

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	require.NoError(t, GitCommit(context.Background(), `feat: handle "quotes" and it's fine`))
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.Dir = repo
	out, err := cmd.Output()
	require.NoError(t, err)
	require.Contains(t, string(out), `feat: handle "quotes" and it's fine`)
}
