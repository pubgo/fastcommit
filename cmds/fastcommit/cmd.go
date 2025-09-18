package fastcommit

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
)

type Config struct {
	GenVersion bool `yaml:"gen_version"`
}

type Params struct {
	Di           *dix.Dix
	Cmd          []*cli.Command
	Cfg          *configs.Config
	OpenaiClient *utils.OpenaiClient
	CommitCfg    []*Config
}

func New(version string) func(params Params) *Command {
	return func(params Params) *Command {
		app := &cli.Command{
			Name:                   "fastcommit",
			Suggest:                true,
			UseShortOptionHandling: true,
			ShellComplete:          cli.DefaultAppComplete,
			Usage:                  "Intelligent generation of git commit message",
			Version:                version,
			Commands:               params.Cmd,
			EnableShellCompletion:  true,
			Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
				if !term.IsTerminal(os.Stdin.Fd()) {
					return ctx, fmt.Errorf("stdin is not terminal")
				}

				if utils.IsHelp() {
					return ctx, cli.ShowAppHelp(command)
				}
				return ctx, nil
			},
			//Action: func(ctx context.Context, command *cli.Command) (gErr error) {
			//	defer result.RecoveryErr(&gErr, func(err error) error {
			//		if errors.Is(err, context.Canceled) {
			//			return nil
			//		}
			//
			//		if err.Error() == "signal: interrupt" {
			//			return nil
			//		}
			//
			//		return err
			//	})
			//
			//	if command.Args().Len() > 0 {
			//		log.Error(ctx).Msgf("unknown command:%v", command.Args().Slice())
			//		cli.ShowRootCommandHelpAndExit(command, 1)
			//		return nil
			//	}
			//
			//	return command.Command("commit").Run(ctx, command.Args().Slice())
			//},
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		return &Command{cmd: app}
	}
}

type Command struct {
	cmd *cli.Command
}

func (c *Command) Run() {
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
	assert.Must(c.cmd.Run(utils.Context(), os.Args))
}
