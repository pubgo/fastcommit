package bootstrap

import (
	"github.com/pubgo/funk/running"
	"os"

	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dixinternal"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/rs/zerolog"
	_ "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/fastcommit/cmds/envcmd"
	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

func syncConfig() {
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

func Main() {
	defer recovery.Exit()

	syncConfig()

	dixinternal.SetLog(func(logger log.Logger) log.Logger {
		if running.IsDebug {
			return logger.WithLevel(zerolog.DebugLevel)
		}
		return logger.WithLevel(zerolog.InfoLevel)
	})

	var di = dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(configs.New)
	di.Provide(tagcmd.New)
	di.Provide(config.Load[ConfigProvider])
	di.Provide(utils.NewOpenaiClient)
	di.Provide(envcmd.New)
	di.Provide(fastcommit.New)
	di.Inject(func(cmd *fastcommit.Command) { cmd.Run() })
}
