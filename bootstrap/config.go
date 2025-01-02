package bootstrap

import "github.com/pubgo/fastcommit/utils"

type Config struct {
	OpenaiConfig *utils.OpenaiConfig `yaml:"openai"`
}
