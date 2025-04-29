package bootstrap

import (
	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"
	_ "github.com/sashabaranov/go-openai"

	"github.com/pubgo/fastcommit/cmds/envcmd"
	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

func Main() {
	defer recovery.Exit()

	initConfig()

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
