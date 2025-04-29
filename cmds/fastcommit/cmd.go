package fastcommit

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/fastcommit/cmds/cmdutils"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

type Params struct {
	Di           *dix.Dix
	Cmd          []*cli.Command
	Cfg          *configs.Config
	OpenaiClient *utils.OpenaiClient
}

func New(params Params) *Command {
	var showPrompt = false
	app := &cli.Command{
		Name:                   "fastcommit",
		Suggest:                true,
		UseShortOptionHandling: true,
		ShellComplete:          cli.DefaultAppComplete,
		Usage:                  "Intelligent generation of git commit message",
		Version:                version.Version(),
		Commands:               params.Cmd,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "show-prompt",
				Usage:       "show prompt",
				Value:       false,
				Destination: &showPrompt,
			},
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			if !term.IsTerminal(os.Stdin.Fd()) {
				return ctx, fmt.Errorf("stdin is not a terminal")
			}

			if utils.IsHelp() {
				return ctx, cli.ShowAppHelp(command)
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			cmdutils.LoadConfigAndBranch()

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
			for _, file := range diff.Files {
				log.Info().Msg("file: " + file)
			}

			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
				s.Prefix = "generate git message: "
			})
			s.Start()
			resp, err := params.OpenaiClient.Client.CreateChatCompletion(
				ctx,
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
			var p1 = tea.NewProgram(InitialTextInputModel(msg))
			mm := assert.Must1(p1.Run()).(model2)
			if mm.isExit() {
				return nil
			}

			msg = mm.Value()
			assert.Must(utils.RunShell("git", "commit", "-m", fmt.Sprintf("'%s'", msg)))
			assert.Must(utils.RunShell("git", "push", "origin", configs.GetBranchName()))
			if showPrompt {
				fmt.Println(generatePrompt)
			}
			log.Info().Any("usage", resp.Usage).Msg("openai response usage")
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
