package configcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "config management",
		Commands: []*cli.Command{
			{
				Name:      "edit",
				Usage:     "edit config, env or local env file, args: [config|env|local], default:config",
				UsageText: "fastcommit config edit [config|env|local]",
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
						if pathutil.IsNotExist(configs.GetLocalEnvPath()) {
							file := assert.Exit1(os.Create(configs.GetLocalEnvPath()))
							defer file.Close()
							for name, cfg := range config.LoadEnvConfigMap(configs.GetConfigPath()) {
								envData := strutil.FirstNotEmpty(cfg.Value, cfg.Default, "")
								fmt.Fprintln(file, fmt.Sprintf(`%s=%q`, name, envData))
							}
						}
						utils.Edit(configs.GetLocalEnvPath())
					}

					return nil
				},
			},

			{
				Name:      "show",
				Usage:     "show config, env or local env file, args: [config|env|local], default:config",
				UsageText: "fastcommit config show [config|env|local]",
				Action: func(ctx context.Context, command *cli.Command) error {
					defer recovery.Exit()

					args := command.Args()
					if args.Len() == 0 || args.First() == "config" {
						cfgPath := configs.GetConfigPath()
						log.Info().Msgf("config path: %s", cfgPath)

						cfgData := assert.Must1(os.ReadFile(cfgPath))
						cfgData = assert.Must1(envsubst.Bytes(cfgData))

						log.Info().Msgf("config data: \n%s", cfgData)
						return nil
					}

					switch args.First() {
					case "env":
						log.Info().Msgf("env path: %s", configs.GetEnvPath())
						env.LoadFiles(configs.GetLocalEnvPath())
						envMap := config.LoadEnvConfigMap(configs.GetConfigPath())
						for name, cfg := range envMap {
							envData := env.Get(name)
							if envData != "" {
								cfg.Value = envData
							}
						}

						pretty.Println(lo.Values(envMap))
					case "local":
						log.Info().Msgf("local env path: %s", configs.GetLocalEnvPath())
						data := result.Wrap(os.ReadFile(configs.GetLocalEnvPath())).Must()
						dataMap := result.Wrap(godotenv.UnmarshalBytes(data)).Must()
						pretty.Println(dataMap)
					}

					return nil
				},
			},
		},
	}
}
