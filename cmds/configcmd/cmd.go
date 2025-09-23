package configcmd

import (
	"context"
	"os"

	"github.com/a8m/envsubst"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pretty"
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

			cfgData := assert.Must1(os.ReadFile(cfgPath))
			cfgData = assert.Must1(envsubst.Bytes(cfgData))

			log.Info().Msgf("config data: \n%s", cfgData)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "edit",
				Usage: "edit config env or local env file, args: [config|env|local]",
				Action: func(ctx context.Context, command *cli.Command) error {
					args := command.Args()
					if args.Len() == 0 {
						utils.Edit(configs.GetConfigPath())
						return nil
					}

					switch args.First() {
					case "config":
						utils.Edit(configs.GetConfigPath())
					case "env":
						utils.Edit(configs.GetEnvPath())
					case "local":
						utils.Edit(configs.GetLocalEnvPath())
					}

					return nil
				},
			},

			{
				Name:  "env",
				Usage: "show all envs",
				Action: func(ctx context.Context, command *cli.Command) error {
					defer recovery.Exit()

					envMap := configs.GetEnvMap()
					for name, cfg := range envMap {
						envData := env.Get(name)
						if envData == "" {
							continue
						}
						cfg.Default = envData
					}

					pretty.Println(lo.Values(envMap))
					return nil
				},
			},
		},
	}
}
