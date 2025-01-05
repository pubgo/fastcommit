package envcmd

import (
	"context"
	
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "env",
		Usage: "show all envs",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			var envData = configs.GetEnvConfig()
			var envMap = make(map[string]*configs.EnvConfig)
			assert.Must(yaml.Unmarshal(envData, &envMap))
			for name := range envMap {
				envMap[name].Name = name
			}

			pretty.Println(lo.Values(envMap))

			return nil
		},
	}
}
