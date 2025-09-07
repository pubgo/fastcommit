package configcmd

import (
	"context"
	"os"

	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
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

			log.Info().Msgf("config data: \n%s", lo.Must(os.ReadFile(cfgPath)))
			return nil
		},
	}
}
