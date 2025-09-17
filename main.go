package main

import (
	_ "embed"

	"github.com/pubgo/fastcommit/bootstrap"
)

//go:embed .version
var version string

func main() {
	bootstrap.Main(version)
}
