package version

var mainPath string

// git rev-parse HEAD
// git describe --always --abbrev=7 --dirty
var (
	commitID  string
	buildTime string
	version   = "v0.0.1-dev-99"
	project   = "project"
)

// git describe --tags --abbrev=0
// git tag --sort=committerdate | tail -n 1
