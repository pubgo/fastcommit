package addcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddArgs(t *testing.T) {
	t.Run("stage all", func(t *testing.T) {
		args, err := addArgs(".")
		require.NoError(t, err)
		require.Equal(t, []string{"add", "."}, args)
	})

	t.Run("stage multiple files", func(t *testing.T) {
		args, err := addArgs("a.go", "b.go")
		require.NoError(t, err)
		require.Equal(t, []string{"add", "a.go", "b.go"}, args)
	})

	t.Run("no files is an error", func(t *testing.T) {
		_, err := addArgs()
		require.ErrorContains(t, err, "usage")
	})
}
