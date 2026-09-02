package fetchcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchArgs(t *testing.T) {
	t.Run("plain fetch", func(t *testing.T) {
		require.Equal(t, []string{"fetch"}, fetchArgs(false))
	})

	t.Run("fetch with prune", func(t *testing.T) {
		require.Equal(t, []string{"fetch", "--prune"}, fetchArgs(true))
	})
}
