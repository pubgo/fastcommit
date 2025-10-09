package version

import (
	_ "embed"
)

//go:embed .version
var version string

// ReleaseVersion v0.0.8
func ReleaseVersion() string { return version }

// ReleaseDate 2025-10-09T12:54:55Z
func ReleaseDate() int64 { return 1760014495 }
