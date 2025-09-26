package version

import (
	_ "embed"
)

//go:embed .version
var version string

func Version() string { return version }

func Date() string { return "2025-09-27" }
