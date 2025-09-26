package bootstrap

import (
	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/recovery"
	_ "github.com/sashabaranov/go-openai"

	"github.com/pubgo/fastcommit/cmds/configcmd"
	"github.com/pubgo/fastcommit/cmds/fastcommit"
	"github.com/pubgo/fastcommit/cmds/fastcommitcmd"
	"github.com/pubgo/fastcommit/cmds/historycmd"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/upgradecmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/utils"
)

func Main() {
	defer recovery.Exit()

	initConfig()

	di := dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(upgradecmd.New)
	di.Provide(tagcmd.New)
	di.Provide(config.Load[configProvider])
	di.Provide(utils.NewOpenaiClient)
	di.Provide(historycmd.New)
	di.Provide(fastcommitcmd.New)
	di.Provide(configcmd.New)
	di.Inject(fastcommit.Run)
}
