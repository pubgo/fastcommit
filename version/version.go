package version

import (
	_ "embed"
)

//go:embed .version
var version string

func Version() string { return version }
