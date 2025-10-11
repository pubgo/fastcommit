package main

import (
	_ "embed"

	"github.com/pubgo/fastcommit/bootstrap"
	"github.com/pubgo/funk/v2/buildinfo/version"
)

//go:embed .version
var release string
var _ = version.SetReleaseVersion(release)

func main() {
	bootstrap.Main()
}
