package fastcommitcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSquashBase(t *testing.T) {
	prefix := "chore: quick update main"

	t.Run("no quick updates returns empty", func(t *testing.T) {
		got := findSquashBase([]string{
			"aaa feat: add feature",
			"bbb fix: bug",
		}, prefix)
		require.Empty(t, got)
	})

	t.Run("squash onto first non-prefix", func(t *testing.T) {
		got := findSquashBase([]string{
			"c1 chore: quick update main at 2026-01-01",
			"c2 chore: quick update main at 2026-01-02",
			"c3 feat: real work",
			"c4 chore: older",
		}, prefix)
		require.Equal(t, "c3", got)
	})

	t.Run("all prefix returns empty", func(t *testing.T) {
		got := findSquashBase([]string{
			"c1 chore: quick update main at a",
			"c2 chore: quick update main at b",
		}, prefix)
		require.Empty(t, got)
	})
}
