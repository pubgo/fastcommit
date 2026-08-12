package aiprovider

import "strings"

// AutoPickCandidate chooses the best commit message without user interaction.
func AutoPickCandidate(candidates []CommitCandidate) string {
	for _, c := range candidates {
		if strings.EqualFold(strings.TrimSpace(c.Style), "conventional") {
			if msg := strings.TrimSpace(c.Message); msg != "" {
				return msg
			}
		}
	}
	for _, c := range candidates {
		if msg := strings.TrimSpace(c.Message); msg != "" {
			return msg
		}
	}
	return ""
}
