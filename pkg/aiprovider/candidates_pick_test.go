package aiprovider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoPickCandidatePrefersConventional(t *testing.T) {
	msg := AutoPickCandidate([]CommitCandidate{
		{Style: "short", Message: "short msg"},
		{Style: "conventional", Message: "feat: add thing"},
		{Style: "medium", Message: "medium message"},
	})
	require.Equal(t, "feat: add thing", msg)
}

func TestAutoPickCandidateFallbackFirst(t *testing.T) {
	msg := AutoPickCandidate([]CommitCandidate{
		{Style: "short", Message: "short msg"},
	})
	require.Equal(t, "short msg", msg)
}
