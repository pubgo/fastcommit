package fastcommit

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"
)

type Params struct {
	Di           *dix.Dix
	Cmd          []*cli.Command
	Cfg          *configs.Config
	OpenaiClient *utils.OpenaiClient
}

func New(params Params) *Command {
	app := &cli.Command{
		Name:                   "fastcommit",
		Suggest:                true,
		UseShortOptionHandling: true,
		ShellComplete:          cli.DefaultAppComplete,
		Usage:                  "Intelligent generation of git commit message",
		Version:                version.Version(),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "enable debug mode",
				Local:       true,
				Value:       running.IsDebug,
				Destination: &running.IsDebug,
				Sources:     cli.EnvVars(env.Key("debug"), env.Key("enable_debug")),
			},
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			branchName := configs.GetBranchName()
			log.Info().Msg("current branch: " + strings.TrimSpace(branchName))
			log.Info().Msg("config: " + configs.GetConfigPath())
			return ctx, nil
		},
		Commands: params.Cmd,
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			if utils.IsHelp() {
				return cli.ShowAppHelp(command)
			}

			if !term.IsTerminal(os.Stdin.Fd()) {
				return nil
			}

			generatePrompt := utils.GeneratePrompt("en", 50, utils.ConventionalCommitType)

			repoPath := assert.Must1(utils.AssertGitRepo())
			log.Info().Msg("git repo: " + repoPath)

			assert.Must(utils.RunShell("git", "add", "--update"))

			diff := assert.Must1(utils.GetStagedDiff(nil))
			if diff == nil {
				return nil
			}

			if len(diff.Files) == 0 {
				return nil
			}

			log.Info().Msg(utils.GetDetectedMessage(diff.Files))

			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
				s.Prefix = "generate git message: "
			})
			s.Start()
			resp, err := params.OpenaiClient.Client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: params.OpenaiClient.Cfg.Model,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: generatePrompt,
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: diff.Diff,
						},
					},
				},
			)
			s.Stop()

			if err != nil {
				log.Err(err).Msg("failed to call openai")
				return errors.WrapCaller(err)
			}

			if len(resp.Choices) == 0 {
				return nil
			}

			msg := resp.Choices[0].Message.Content
			log.Info().Str("msg", msg).Msg("openai response git message")
			var p1 = tea.NewProgram(InitialTextInputModel(msg))
			mm := assert.Must1(p1.Run()).(model2)
			if mm.isExit() {
				return nil
			}

			msg = mm.Value()
			assert.Must(utils.RunShell("git", "commit", "-m", fmt.Sprintf("'%s'", msg)))
			assert.Must(utils.RunShell("git", "push", "origin", configs.GetBranchName()))

			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	return &Command{cmd: app}
}

type Command struct {
	cmd *cli.Command
}

func (c *Command) Run() {
	defer recovery.Exit()
	assert.Must(c.cmd.Run(utils.Context(), os.Args))
}
