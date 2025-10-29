package bootstrap

import (
	"context"
	"fmt"
	"os"

	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixcontext"
	"github.com/pubgo/fastcommit/cmds/pullcmd"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/features/featureflags"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/samber/lo"
	_ "github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/fastcommit/cmds/configcmd"
	"github.com/pubgo/fastcommit/cmds/fastcommitcmd"
	"github.com/pubgo/fastcommit/cmds/historycmd"
	"github.com/pubgo/fastcommit/cmds/tagcmd"
	"github.com/pubgo/fastcommit/cmds/upgradecmd"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/utils"
)

func Main() {
	run(
		versioncmd.New(),
		upgradecmd.New(),
		tagcmd.New(),
		historycmd.New(),
		fastcommitcmd.New(),
		configcmd.New(),
		pullcmd.New(),
	)
}

func run(cmds ...*cli.Command) {
	defer recovery.Exit(func(err error) error {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		if err.Error() == "signal: interrupt" {
			return nil
		}

		log.Err(err).Msg("failed to run command")
		return nil
	})

	app := &cli.Command{
		Name:                   "fastcommit",
		Suggest:                true,
		UseShortOptionHandling: true,
		ShellComplete:          cli.DefaultAppComplete,
		Usage:                  "Intelligent generation of git commit message",
		Version:                version.ReleaseVersion(),
		Commands:               cmds,
		EnableShellCompletion:  true,
		Flags:                  append(featureflags.GetFlags(), lo.ToPtr(running.DebugFlag)),
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			if !term.IsTerminal(os.Stdin.Fd()) {
				return ctx, fmt.Errorf("stdin is not terminal")
			}

			if utils.IsHelp() {
				return ctx, cli.ShowAppHelp(command)
			}

			initConfig()
			di := dix.New(dix.WithValuesNull())
			di.Provide(config.Load[configProvider])
			di.Provide(utils.NewOpenaiClient)
			return dixcontext.Create(ctx, di), nil
		},
	}

	assert.Must(app.Run(utils.Context(), os.Args))
}
