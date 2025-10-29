package main

import (
	"fmt"
	"os"

	"github.com/pubgo/funk/v2/assert"
	flag "github.com/spf13/pflag"
)

var ip = flag.Int("flagname", 1234, "help message for flagname")

func main() {
	var commandLine = flag.NewFlagSet("test-flag", flag.ExitOnError)
	commandLine.Int("flagname", 1234, "help message for flagname")
	commandLine.StringArray("names", nil, "help message for flagname")
	fmt.Println(commandLine.Args())
	commandLine.PrintDefaults()
	assert.Must(commandLine.Parse(os.Args))
}
