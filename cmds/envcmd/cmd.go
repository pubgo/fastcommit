package envcmd

import (
	"context"

	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/fastcommit/configs"
)

func New() *cli.Command {
	return &cli.Command{
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
	}
}
