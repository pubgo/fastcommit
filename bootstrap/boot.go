package bootstrap

import (
	"context"
	"fmt"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/funk/pretty"
	"github.com/sashabaranov/go-openai"
	"os"
	"sort"

	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_internal"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/rs/zerolog"
	_ "github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"
)

func Main() {
	defer recovery.Exit()

	dix_internal.SetLogLevel(zerolog.InfoLevel)
	var di = dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(config.Load[Config])

	di.Inject(func(cmd []*cli.Command) {
		app := &cli.Command{
			Name:                   "fastcommit",
			Suggest:                true,
			UseShortOptionHandling: true,
			ShellComplete:          cli.DefaultAppComplete,
			Usage:                  "Intelligent generation of git commit message",
			Version:                version.Version(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:       "config",
					Aliases:    []string{"c"},
					Usage:      "config file path",
					Value:      config.GetConfigPath(),
					Persistent: true,
					Sources:    cli.EnvVars(env.Key("fast_commit_config")),
					Action: func(ctx context.Context, command *cli.Command, s string) error {
						config.SetConfigPath(s)
						return nil
					},
				},
				&cli.BoolFlag{
					Name:        "debug",
					Usage:       "enable debug mode",
					Persistent:  true,
					Value:       running.IsDebug,
					Destination: &running.IsDebug,
					Sources:     cli.EnvVars(env.Key("debug"), env.Key("enable_debug")),
				},
			},
			Commands: cmd,
			Action: func(ctx context.Context, command *cli.Command) error {
				if utils.IsHelp() {
					return cli.ShowAppHelp(command)
				}

				fmt.Println(term.IsTerminal(os.Stdin.Fd()))

				pp := utils.GeneratePrompt("en", 50, utils.ConventionalCommitType)

				repoPath, err := utils.AssertGitRepo()
				fmt.Println("Git repository root:", repoPath, err)

				//await execa('git', ['add', '--update']);
				diff, err := utils.GetStagedDiff(nil)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				if diff != nil {
					fmt.Println("Staged files:", diff["files"])
					//fmt.Println("Staged diff:", diff["diff"])
					//fmt.Println(utils.GetDetectedMessage(diff["files"].([]string)))
				}

				var cfg = openai.DefaultConfig("sk-a6af4bae0ef441b299f4301c3feaedf4")
				cfg.BaseURL = "https://api.deepseek.com/v1"
				client := openai.NewClientWithConfig(cfg)

				resp, err := client.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model: "deepseek-chat",
						Messages: []openai.ChatCompletionMessage{
							{
								Role:    openai.ChatMessageRoleSystem,
								Content: pp,
							},
							{
								Role:    openai.ChatMessageRoleUser,
								Content: diff["diff"].(string),
							},
						},
					},
				)

				if err != nil {
					fmt.Printf("ChatCompletion error: %v\n", err)
				}

				pretty.Println(resp)

				return nil
			},
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		assert.Must(app.Run(utils.Context(), os.Args))
	})
}
