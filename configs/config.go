package configs

import (
	_ "embed"
	"path"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/bitfield/script"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/env"
)

const DebugEnvKey = "ENABLE_DEBUG"

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

func GetDefaultConfig() []byte { return defaultConfig }

func GetEnvConfig() []byte { return envConfig }

var IsDebug = sync.OnceValue(func() bool { return env.GetBool(DebugEnvKey) })
