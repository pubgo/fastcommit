package diffcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffArgs(t *testing.T) {
	t.Run("default diffs against HEAD", func(t *testing.T) {
		args, err := diffArgs(false, false)
		require.NoError(t, err)
		require.Equal(t, []string{"diff", "HEAD"}, args)
	})

	t.Run("staged diff", func(t *testing.T) {
		args, err := diffArgs(true, false)
		require.NoError(t, err)
		require.Equal(t, []string{"diff", "--cached"}, args)
	})

	t.Run("unstaged diff", func(t *testing.T) {
		args, err := diffArgs(false, true)
		require.NoError(t, err)
		require.Equal(t, []string{"diff"}, args)
	})

	t.Run("staged and unstaged together is an error", func(t *testing.T) {
		_, err := diffArgs(true, true)
		require.Error(t, err)
	})
}
