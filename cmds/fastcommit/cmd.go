package fastcommit

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/v2/result"
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

func New(version string) func(params Params) *Command {
	return func(params Params) *Command {
		var flags = new(struct {
			showPrompt bool
			fastCommit bool
		})
		app := &cli.Command{
			Name:                   "fastcommit",
			Suggest:                true,
			UseShortOptionHandling: true,
			ShellComplete:          cli.DefaultAppComplete,
			Usage:                  "Intelligent generation of git commit message",
			Version:                version,
			Commands:               params.Cmd,
			EnableShellCompletion:  true,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "prompt",
					Usage:       "show prompt",
					Value:       flags.showPrompt,
					Destination: &flags.showPrompt,
				},
				&cli.BoolFlag{
					Name:        "fast",
					Usage:       "quickly generate messages without prompts",
					Value:       flags.fastCommit,
					Destination: &flags.fastCommit,
				},
			},
			Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
				if !term.IsTerminal(os.Stdin.Fd()) {
					return ctx, fmt.Errorf("stdin is not terminal")
				}

				if utils.IsHelp() {
					return ctx, cli.ShowAppHelp(command)
				}
				return ctx, nil
			},
			Action: func(ctx context.Context, command *cli.Command) (gErr error) {
				defer result.Recovery(&gErr, func(err error) error {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					return err
				})

				defer func() {
					if errors.Is(gErr, context.Canceled) {
						gErr = nil
					}
				}()

				if command.Args().Len() > 0 {
					log.Error(ctx).Msgf("unknown command:%v", command.Args().Slice())
					cli.ShowRootCommandHelpAndExit(command, 1)
					return nil
				}

				isDirty := utils.IsDirty().Must()
				if !isDirty {
					return
				}

				cmdutils.LoadConfigAndBranch()

				allTags := utils.GetAllGitTags()
				tagName := "v0.0.1"
				if len(allTags) > 0 {
					ver := utils.GetNextReleaseTag(allTags)
					tagName = "v" + strings.TrimPrefix(ver.Original(), "v")
				}
				assert.Exit(os.WriteFile(".version", []byte(tagName), 0644))

				repoPath := assert.Must1(utils.AssertGitRepo())
				log.Info().Msg("git repo: " + repoPath)

				username := strings.TrimSpace(assert.Must1(utils.RunOutput("git", "config", "get", "user.name")))

				if flags.fastCommit {
					assert.Must(utils.RunShell("git", "add", "-A"))

					msg := fmt.Sprintf("chore: @%s quick update %s at %s", username, cmdutils.GetBranchName(), time.Now().Format(time.DateTime))
					assert.Must(utils.RunShell("git", "commit", "-m", strconv.Quote(msg)))
					assert.Must(utils.RunShell("git", "push", "origin", cmdutils.GetBranchName()))
					return
				}

				assert.Must(utils.RunShell("git", "add", "--update"))

				diff := assert.Must1(utils.GetStagedDiff(nil))
				if diff == nil || len(diff.Files) == 0 {
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
				generatePrompt := utils.GeneratePrompt("en", 50, utils.ConventionalCommitType)
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
				assert.Must(utils.RunShell("git", "commit", "-m", strconv.Quote(msg)))
				assert.Must(utils.RunShell("git", "push", "origin", cmdutils.GetBranchName()))
				if flags.showPrompt {
					fmt.Println("\n" + generatePrompt + "\n")
				}
				log.Info().Any("usage", resp.Usage).Msg("openai response usage")
				return nil
			},
		}

		sort.Sort(cli.FlagsByName(app.Flags))
		return &Command{cmd: app}
	}
}

type Command struct {
	cmd *cli.Command
}

func (c *Command) Run() {
	defer recovery.Exit(func(err error) error {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})
	assert.Must(c.cmd.Run(utils.Context(), os.Args))
}
