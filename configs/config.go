package configs

import (
	_ "embed"

	"github.com/adrg/xdg"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
)

type Config struct {
	BranchName string
}

type Version struct {
	Name string `yaml:"name"`
}

var configPath = assert.Exit1(xdg.ConfigFile("fastcommit/config.yaml"))
var branchName = assert.Exit1(utils.RunOutput("git", "rev-parse", "--abbrev-ref", "HEAD"))

//go:embed default.yaml
var defaultConfig []byte

func GetConfigPath() string {
	return configPath
}

func GetBranchName() string {
	return branchName
}

func GetDefaultConfig() []byte {
	return defaultConfig
}
