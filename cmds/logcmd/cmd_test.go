package logcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogArgs(t *testing.T) {
	t.Run("simple log", func(t *testing.T) {
		require.Equal(t, []string{"log", "--oneline", "-20"}, logArgs(false))
	})

	t.Run("graph log", func(t *testing.T) {
		require.Equal(t, []string{"log", "--graph", "--decorate", "--oneline", "-30"}, logArgs(true))
	})
}
