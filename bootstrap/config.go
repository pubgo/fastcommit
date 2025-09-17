package bootstrap

import (
	"log/slog"
	"os"

	"github.com/pubgo/dix/dixinternal"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

type ConfigProvider struct {
	Version      *configs.Version    `yaml:"version"`
	OpenaiConfig *utils.OpenaiConfig `yaml:"openai"`
	CommitConfig *fastcommit.Config  `yaml:"commit"`
}

func initConfig() {
	slog.SetDefault(slog.New(log.NewSlog(log.GetLogger("fastcommit"))))

	configs.InitEnv()

	dixinternal.SetLog(func(logger log.Logger) log.Logger {
		if configs.IsDebug() {
			return logger.WithLevel(zerolog.DebugLevel)
		}
		return logger.WithLevel(zerolog.ErrorLevel)
	})

	configPath := configs.GetConfigPath()
	defaultConfigData := configs.GetDefaultConfig()
	if pathutil.IsNotExist(configPath) {
		assert.Must(os.WriteFile(configPath, defaultConfigData, 0644))
		return
	}

	var cfg ConfigProvider
	config.LoadFromPath(&cfg, configPath)

	var defaultCfg ConfigProvider
	assert.Must(yaml.Unmarshal(defaultConfigData, &defaultCfg))
	if cfg.Version == nil || cfg.Version.Name == "" || defaultCfg.Version.Name != cfg.Version.Name {
		assert.Must(os.WriteFile(configPath, defaultConfigData, 0644))
	}

	config.SetConfigPath(configPath)
}
