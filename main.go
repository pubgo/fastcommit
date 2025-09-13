package main

import (
	_ "embed"

	_ "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pubgo/fastcommit/bootstrap"
)

//go:embed .version
var version string

func main() {
	bootstrap.Main(version)
}
