package bootstrap

import (
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

type ConfigProvider struct {
	Version      *configs.Version    `yaml:"version"`
	OpenaiConfig *utils.OpenaiConfig `yaml:"openai"`
}
