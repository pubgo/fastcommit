package statuscmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusArgs(t *testing.T) {
	t.Run("full status", func(t *testing.T) {
		require.Equal(t, []string{"status"}, statusArgs(false))
	})

	t.Run("short status", func(t *testing.T) {
		require.Equal(t, []string{"status", "--short"}, statusArgs(true))
	})
}
