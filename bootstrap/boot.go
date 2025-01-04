package bootstrap

import (
	"log/slog"
	"os"

	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_internal"
	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/typex"
	"github.com/rs/zerolog"
	_ "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

func Main() {
	defer recovery.Exit()

	slog.Info("config path", "path", configPath)
	typex.DoBlock(func() {
		if pathutil.IsNotExist(configPath) {
			assert.Must(os.WriteFile(configPath, defaultConfig, 0644))
			return
		}

		var cfg ConfigProvider
		config.LoadFromPath(&cfg, configPath)

		var defaultCfg ConfigProvider
		assert.Exit(yaml.Unmarshal(defaultConfig, &defaultCfg))
		if cfg.Version == nil || cfg.Version.Name == "" || defaultCfg.Version.Name != cfg.Version.Name {
			assert.Exit(os.WriteFile(configPath, defaultConfig, 0644))
		}
	})

	config.SetConfigPath(configPath)
	dix_internal.SetLogLevel(zerolog.InfoLevel)

	var di = dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(func() *configs.Config {
		return &configs.Config{
			BranchName: branchName,
		}
	})
	di.Provide(tagcmd.New)
	di.Provide(config.Load[ConfigProvider])
	di.Provide(utils.NewOpenaiClient)
	di.Provide(fastcommit.New)
	di.Inject(func(cmd *fastcommit.Command) { cmd.Run() })
}
