package rebasecmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRebaseArgs(t *testing.T) {
	t.Run("rebase onto upstream", func(t *testing.T) {
		args, err := rebaseArgs("main", false, false, false)
		require.NoError(t, err)
		require.Equal(t, []string{"rebase", "main"}, args)
	})

	t.Run("continue in-progress rebase", func(t *testing.T) {
		args, err := rebaseArgs("", true, false, false)
		require.NoError(t, err)
		require.Equal(t, []string{"rebase", "--continue"}, args)
	})

	t.Run("abort in-progress rebase", func(t *testing.T) {
		args, err := rebaseArgs("", false, true, false)
		require.NoError(t, err)
		require.Equal(t, []string{"rebase", "--abort"}, args)
	})

	t.Run("skip current patch", func(t *testing.T) {
		args, err := rebaseArgs("", false, false, true)
		require.NoError(t, err)
		require.Equal(t, []string{"rebase", "--skip"}, args)
	})

	t.Run("no upstream and no flag is an error", func(t *testing.T) {
		_, err := rebaseArgs("", false, false, false)
		require.ErrorContains(t, err, "usage")
	})

	t.Run("upstream with control flag is an error", func(t *testing.T) {
		_, err := rebaseArgs("main", true, false, false)
		require.Error(t, err)
	})

	t.Run("multiple control flags are an error", func(t *testing.T) {
		_, err := rebaseArgs("", true, true, false)
		require.Error(t, err)
	})
}
