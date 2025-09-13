package bootstrap

import (
	"log/slog"

	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	_ "github.com/sashabaranov/go-openai"

	"github.com/pubgo/fastcommit/cmds/configcmd"
	"github.com/pubgo/fastcommit/cmds/envcmd"
	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/cmds/historycmd"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/upgradecmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

func Main(ver string) {
	defer recovery.Exit()

	slog.SetDefault(slog.New(log.NewSlog(log.GetLogger("fastcommit"))))

	initConfig()

	var di = dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(upgradecmd.New)
	di.Provide(configs.New)
	di.Provide(tagcmd.New)
	di.Provide(config.Load[ConfigProvider])
	di.Provide(utils.NewOpenaiClient)
	di.Provide(envcmd.New)
	di.Provide(historycmd.New)
	di.Provide(fastcommit.New(ver))
	di.Provide(configcmd.New)
	di.Inject(func(cmd *fastcommit.Command) { cmd.Run() })
}
