package configs

import (
	_ "embed"
	"path"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/bitfield/script"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/env"
	"gopkg.in/yaml.v3"
)

const DebugEnvKey = "ENABLE_DEBUG"

type EnvConfig struct {
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Name        string `yaml:"name"`
	Required    bool   `yaml:"required"`
}

type Version struct {
	Name string `yaml:"name"`
}

//go:embed default.yaml
var defaultConfig []byte

//go:embed env.yaml
var envConfig []byte

var GetConfigPath = sync.OnceValue(func() string {
	return assert.Exit1(xdg.ConfigFile("fastcommit/config.yaml"))
})

var GetRepoPath = sync.OnceValue(func() string {
	repoPath := assert.Exit1(script.Exec("git rev-parse --show-toplevel").String())
	return strings.TrimSpace(repoPath)
})

var GetEnvPath = sync.OnceValue(func() string {
	return path.Join(path.Dir(GetConfigPath()), "env.yaml")
})

var GetLocalEnvPath = sync.OnceValue(func() string {
	return path.Join(GetRepoPath(), ".git", "fastcommit.env")
})

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

		env.Set(name, cfg.Default).Must()
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
	env.GetBoolVal(&debug, DebugEnvKey)
	return
}
