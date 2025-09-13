package configcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitfield/script"
	"github.com/pkg/browser"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "config management",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			cfgPath := configs.GetConfigPath()
			log.Info().Msgf("config path: %s", cfgPath)

			log.Info().Msgf("config data: \n%s", assert.Must1(os.ReadFile(cfgPath)))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "edit",
				Usage: "fastcommit config edit [open|vim|zed|code|...]",
				Action: func(ctx context.Context, command *cli.Command) error {
					log.Info().Msgf("config path: %s", configs.GetConfigPath())

					if command.Args().Len() == 0 {
						return browser.OpenFile(configs.GetConfigPath())
					}

					cmd := command.Args().First()

					path := assert.Exit1(filepath.Abs(configs.GetConfigPath()))
					shell := fmt.Sprintf(`%s "%s"`, cmd, path)
					log.Info().Msgf("edit config: %s", shell)
					_, err := script.Exec(shell).Stdout()
					return err
				},
			},
		},
	}
}
