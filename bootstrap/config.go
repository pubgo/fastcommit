package bootstrap

import (
	_ "embed"

	"github.com/adrg/xdg"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
)

type ConfigProvider struct {
	Version      *configs.Version    `yaml:"version"`
	OpenaiConfig *utils.OpenaiConfig `yaml:"openai"`
}

var configPath = assert.Exit1(xdg.ConfigFile("fastcommit/config.yaml"))
var branchName = assert.Exit1(utils.RunOutput("git", "rev-parse", "--abbrev-ref", "HEAD"))

//go:embed default.yaml
var defaultConfig []byte
