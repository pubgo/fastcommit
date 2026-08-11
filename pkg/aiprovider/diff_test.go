package aiprovider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactDiffForAISmallUnchanged(t *testing.T) {
	diff := "diff --git a/a.go b/a.go\n+++ b/a.go\n+package a\n"
	out, stats := CompactDiffForAI(diff)
	require.Equal(t, strings.TrimSpace(diff), out)
	require.False(t, stats.Truncated)
}

func TestCompactDiffForAITruncatesLarge(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 60; i++ {
		name := fmt.Sprintf("pkg/file_%02d.go", i)
		fmt.Fprintf(&b, "diff --git a/%s b/%s\n", name, name)
		b.WriteString(strings.Repeat("+line content for testing truncation\n", 200))
	}
	out, stats := CompactDiffForAI(b.String())
	require.True(t, stats.Truncated)
	require.LessOrEqual(t, len(out), defaultMaxDiffChars+4000)
	require.Contains(t, out, "abbreviated")
	require.Greater(t, stats.SkippedFiles, 0)
	require.LessOrEqual(t, stats.KeptFiles, defaultMaxFiles)
}

func TestCompactDiffForAISkipsLockfiles(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/go.sum b/go.sum\n+++ b/go.sum\n+" + strings.Repeat("h1:abc\n", 100),
		"diff --git a/main.go b/main.go\n+++ b/main.go\n+package main\n",
	}, "\n")
	out, stats := CompactDiffForAI(diff)
	require.NotContains(t, out, strings.Repeat("h1:abc", 20))
	require.Contains(t, out, "main.go")
	require.Contains(t, out, "go.sum (skipped)")
	require.Equal(t, 1, stats.KeptFiles)
}
