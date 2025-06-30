package configs

import (
	_ "embed"

	"github.com/adrg/xdg"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/env"
	"gopkg.in/yaml.v3"
)

const debugEnv = "ENABLE_DEBUG"

type EnvConfig struct {
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Name        string `yaml:"name"`
	Required    bool   `yaml:"required"`
}

func New() *Config {
	return &Config{}
}

type Config struct {
}

type Version struct {
	Name string `yaml:"name"`
}

var configPath string
var branchName string

//go:embed default.yaml
var defaultConfig []byte

//go:embed env.yaml
var envConfig []byte

func GetConfigPath() string {
	if configPath != "" {
		return configPath
	}

	configPath = assert.Exit1(xdg.ConfigFile("fastcommit/config.yaml"))
	return configPath
}

func GetDefaultConfig() []byte {
	return defaultConfig
}

func GetEnvConfig() []byte {
	return envConfig
}

func InitEnv() {
	envMap := GetEnvMap()
	for name, cfg := range envMap {
		envData := env.Get(name)
		if envData == "" {
			continue
		}
		cfg.Default = envData
	}

	for name, cfg := range envMap {
		if cfg.Required && cfg.Default == "" {
			panic("env " + cfg.Name + " is required")
		}

		assert.Must(env.Set(name, cfg.Default))
	}
}

func GetEnvMap() map[string]*EnvConfig {
	var envData = GetEnvConfig()
	var envMap = make(map[string]*EnvConfig)
	assert.Must(yaml.Unmarshal(envData, &envMap))
	for name := range envMap {
		envMap[name].Name = name
	}
	return envMap
}

func IsDebug() (debug bool) {
	env.GetBoolVal(&debug, debugEnv)
	return
}
