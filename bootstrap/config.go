package bootstrap

import (
	"context"
	"github.com/pubgo/fastcommit/cmds/fastcommitcmd"
	"log/slog"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

type configProvider struct {
	Version      *configs.Version      `yaml:"version"`
	OpenaiConfig *utils.OpenaiConfig   `yaml:"openai"`
	CommitConfig *fastcommitcmd.Config `yaml:"commit"`
}

func initConfig() {
	env.Set("LC_ALL", "C").Must()
	slog.SetDefault(slog.New(log.NewSlog(log.GetLogger("fastcommit"))))
	log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Map) bool {
		if configs.IsDebug() {
			return true
		}

		if name == "dix" || name == "env" || fields["module"] == "env" {
			return false
		}
		return true
	})

	env.LoadFiles(configs.GetLocalEnvPath())
	configs.InitEnv()

	configPath := configs.GetConfigPath()
	envPath := configs.GetEnvPath()
	if pathutil.IsNotExist(configPath) {
		assert.Must(os.WriteFile(configPath, configs.GetDefaultConfig(), 0644))
		assert.Must(os.WriteFile(envPath, configs.GetEnvConfig(), 0644))
		return
	}

	type versionConfigProvider struct {
		Version *configs.Version `yaml:"version"`
	}
	var cfg versionConfigProvider
	config.LoadFromPath(&cfg, configPath)

	var defaultCfg versionConfigProvider
	defaultConfigData := configs.GetDefaultConfig()
	assert.Must(yaml.Unmarshal(defaultConfigData, &defaultCfg))
	if cfg.Version == nil || cfg.Version.Name == "" || defaultCfg.Version.Name != cfg.Version.Name {
		assert.Must(os.WriteFile(configPath, defaultConfigData, 0644))
		assert.Must(os.WriteFile(envPath, configs.GetEnvConfig(), 0644))
	}

	config.SetConfigPath(configPath)
}
