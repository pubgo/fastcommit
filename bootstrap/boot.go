package bootstrap

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/adrg/xdg"
	_ "github.com/adrg/xdg"
	_ "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_internal"
	"github.com/pubgo/fastcommit/cmds/versioncmd"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/rs/zerolog"
	"github.com/sashabaranov/go-openai"
	_ "github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"
)

var configPath = assert.Exit1(xdg.ConfigFile("fastcommit/config.yaml"))

//go:embed default.yaml
var defaultConfig []byte

func Main() {
	defer recovery.Exit()

	var branchName = string(assert.Exit1(utils.ShellOutput("git", "rev-parse", "--abbrev-ref", "HEAD")))

	slog.Info("config path", "path", configPath)
	if pathutil.IsNotExist(configPath) {
		assert.Must(os.WriteFile(configPath, defaultConfig, 0644))
	}

	config.SetConfigPath(configPath)

	dix_internal.SetLogLevel(zerolog.InfoLevel)
	var di = dix.New(dix.WithValuesNull())
	di.Provide(versioncmd.New)
	di.Provide(config.Load[Config])
	di.Provide(utils.NewOpenaiClient)

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

				if !term.IsTerminal(os.Stdin.Fd()) {
					return nil
				}

				generatePrompt := utils.GeneratePrompt("en", 50, utils.ConventionalCommitType)

				repoPath := assert.Must1(utils.AssertGitRepo())
				slog.Info("Git repository root", "path", repoPath)

				assert.Exit(utils.Shell("git", "add", "--update").Run())

				diff := assert.Must1(utils.GetStagedDiff(nil))

				client := dix.Inject(di, new(struct {
					*utils.OpenaiClient
				}))
				resp, err := client.Client.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model: client.Cfg.Model,
						Messages: []openai.ChatCompletionMessage{
							{
								Role:    openai.ChatMessageRoleSystem,
								Content: generatePrompt,
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

				if len(resp.Choices) == 0 {
					return nil
				}

				msg := resp.Choices[0].Message.Content
				assert.Must(utils.Shell("git", "commit", "-m", fmt.Sprintf("'%s'", msg)).Run())
				assert.Must(utils.Shell("git", "push", "origin", branchName).Run())

				return nil
			},
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		assert.Must(app.Run(utils.Context(), os.Args))
	})
}
