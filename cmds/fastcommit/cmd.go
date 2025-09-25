package fastcommit

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/x/term"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/fastcommit/version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
)

type params struct {
	Cmds []*cli.Command
}

func Run(params params) {
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
		Version:                version.Version(),
		Commands:               params.Cmds,
		EnableShellCompletion:  true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "enable debug",
				Value:   false,
				Sources: cli.EnvVars(configs.DebugEnvKey),
				Action: func(ctx context.Context, command *cli.Command, b bool) error {
					return os.Setenv(configs.DebugEnvKey, "true")
				},
			},
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			if !term.IsTerminal(os.Stdin.Fd()) {
				return ctx, fmt.Errorf("stdin is not terminal")
			}

			if utils.IsHelp() {
				return ctx, cli.ShowAppHelp(command)
			}
			return ctx, nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	assert.Must(app.Run(utils.Context(), os.Args))
}
