package aiprovider

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	defaultMaxDiffChars = 24_000
	defaultMaxFileChars = 4_000
	defaultMaxFiles     = 40
)

// DiffCompactStats describes how a diff was reduced for AI prompts.
type DiffCompactStats struct {
	OriginalBytes int
	CompactBytes  int
	FileCount     int
	KeptFiles     int
	SkippedFiles  int
	Truncated     bool
}

// CompactDiffForAI shrinks a git diff so commit/review prompts stay within model limits.
// It keeps a file list summary and truncated hunks, skipping lockfiles and binary patches.
func CompactDiffForAI(diff string) (string, DiffCompactStats) {
	diff = strings.TrimSpace(diff)
	stats := DiffCompactStats{OriginalBytes: len(diff)}
	if diff == "" {
		return "", stats
	}
	if len(diff) <= defaultMaxDiffChars && !looksLikeHugeSingleFile(diff) && !hasSkippableDiffPaths(diff) {
		stats.CompactBytes = len(diff)
		stats.FileCount = countDiffFiles(diff)
		stats.KeptFiles = stats.FileCount
		return diff, stats
	}

	sections := splitDiffSections(diff)
	stats.FileCount = len(sections)

	var kept []string
	var omitted []string
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Staged changes summary (%d files). Diff abbreviated for AI.\n\n", len(sections)))

	for _, section := range sections {
		path := diffSectionPath(section)
		if path == "" {
			path = "(unknown)"
		}
		if shouldSkipDiffPath(path) || isBinaryDiffSection(section) {
			omitted = append(omitted, path+" (skipped)")
			continue
		}
		if len(kept) >= defaultMaxFiles {
			omitted = append(omitted, path)
			continue
		}

		body := section
		if len(body) > defaultMaxFileChars {
			body = body[:defaultMaxFileChars] + "\n... (file diff truncated)\n"
			stats.Truncated = true
		}
		if b.Len()+len(body)+1 > defaultMaxDiffChars {
			omitted = append(omitted, path)
			stats.Truncated = true
			continue
		}
		if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
			b.WriteByte('\n')
		}
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
		kept = append(kept, path)
	}

	stats.KeptFiles = len(kept)
	stats.SkippedFiles = len(omitted)
	if len(omitted) > 0 {
		stats.Truncated = true
		b.WriteString("\nOmitted files:\n")
		limit := len(omitted)
		if limit > 30 {
			limit = 30
		}
		for _, name := range omitted[:limit] {
			b.WriteString("- ")
			b.WriteString(name)
			b.WriteByte('\n')
		}
		if len(omitted) > limit {
			fmt.Fprintf(&b, "- ... and %d more\n", len(omitted)-limit)
		}
	}

	out := strings.TrimSpace(b.String())
	stats.CompactBytes = len(out)
	if stats.CompactBytes < stats.OriginalBytes {
		stats.Truncated = true
	}
	return out, stats
}

func hasSkippableDiffPaths(diff string) bool {
	for _, section := range splitDiffSections(diff) {
		path := diffSectionPath(section)
		if shouldSkipDiffPath(path) || isBinaryDiffSection(section) {
			return true
		}
	}
	return false
}

func looksLikeHugeSingleFile(diff string) bool {
	return countDiffFiles(diff) <= 1 && len(diff) > defaultMaxFileChars
}

func countDiffFiles(diff string) int {
	return strings.Count(diff, "diff --git ")
}

func splitDiffSections(diff string) []string {
	parts := strings.Split(diff, "diff --git ")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, "diff --git "+part)
	}
	return out
}

func diffSectionPath(section string) string {
	// diff --git a/path b/path
	line, _, _ := strings.Cut(section, "\n")
	fields := strings.Fields(line)
	if len(fields) >= 4 {
		path := strings.TrimPrefix(fields[3], "b/")
		if path != "" {
			return path
		}
	}
	if len(fields) >= 3 {
		return strings.TrimPrefix(fields[2], "a/")
	}
	return ""
}

func isBinaryDiffSection(section string) bool {
	lower := strings.ToLower(section)
	return strings.Contains(lower, "binary files ") ||
		strings.Contains(lower, "git binary patch")
}

func shouldSkipDiffPath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "go.sum", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "composer.lock", "poetry.lock", "cargo.lock":
		return true
	}
	if strings.HasSuffix(base, ".min.js") || strings.HasSuffix(base, ".min.css") {
		return true
	}
	if strings.HasSuffix(base, ".map") || strings.HasSuffix(base, ".wasm") {
		return true
	}
	if strings.Contains(path, "node_modules/") || strings.Contains(path, "vendor/") {
		return true
	}
	return false
}
