package fastcommitcmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/x/term"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/v2/result"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/tap"

	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
)

type Config struct {
	GenVersion bool `yaml:"gen_version"`
}

type Params struct {
	Di           *dix.Dix
	Cfg          *configs.Config
	OpenaiClient *utils.OpenaiClient
	CommitCfg    []*Config
}

func New(params Params) *cli.Command {
	var flags = new(struct {
		showPrompt bool
		fastCommit bool
	})
	app := &cli.Command{
		Name:                   "commit",
		Suggest:                true,
		UseShortOptionHandling: true,
		ShellComplete:          cli.DefaultAppComplete,
		Usage:                  "Intelligent generation of git commit message",
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
			defer result.RecoveryErr(&gErr, func(err error) error {
				if errors.Is(err, context.Canceled) {
					return nil
				}

				if err.Error() == "signal: interrupt" {
					return nil
				}

				return err
			})

			if command.Args().Len() > 0 {
				log.Error(ctx).Msgf("unknown command:%v", command.Args().Slice())
				cli.ShowRootCommandHelpAndExit(command, 1)
				return nil
			}

			utils.LoadConfigAndBranch()

			err := utils.PreGitPush(ctx)
			if err != nil {
				if shouldPullDueToRemoteUpdate(err.Error()) {
					err := gitPull()
					if err != nil {
						if isMergeConflict() {
							handleMergeConflict()
						} else {
							os.Exit(1)
						}
					} else {
						informUserToAmendAndPush()
					}
				}
			}

			isDirty := utils.IsDirty().Must()
			if !isDirty {
				return
			}

			for _, cfg := range params.CommitCfg {
				if !cfg.GenVersion {
					continue
				}

				allTags := utils.GetAllGitTags(ctx)
				tagName := "v0.0.1"
				if len(allTags) > 0 {
					ver := utils.GetNextReleaseTag(allTags)
					tagName = "v" + strings.TrimPrefix(ver.Original(), "v")
				}
				assert.Must(pathutil.IsNotExistMkDir("version"))
				assert.Exit(os.WriteFile("version/.version", []byte(tagName), 0644))
				assert.Exit(os.WriteFile("version/version.go", []byte(`package version

import (
	_ "embed"
)

//go:embed .version
var version string

func Version() string { return version }
`), 0644))
				break
			}

			repoPath := assert.Must1(utils.AssertGitRepo(ctx))
			log.Info().Msg("git repo: " + repoPath)

			//username := strings.TrimSpace(assert.Must1(utils.RunOutput("git", "config", "get", "user.name")))

			if flags.fastCommit {
				preMsg := strings.TrimSpace(assert.Must1(utils.RunOutput(ctx, "git", "log", "-1", "--pretty=%B")))
				prefixMsg := fmt.Sprintf("chore: quick update %s", utils.GetBranchName())
				msg := fmt.Sprintf("%s at %s", prefixMsg, time.Now().Format(time.DateTime))

				msg = strings.TrimSpace(tap.Text(ctx, tap.TextOptions{
					Message:      "git message(update or enter):",
					InitialValue: msg,
					DefaultValue: msg,
					Placeholder:  "update or enter",
				}))

				if msg == "" {
					return
				}

				assert.Must(utils.RunShell(ctx, "git", "add", "-A"))
				res := assert.Must1(utils.RunOutput(ctx, "git", "status"))
				if strings.Contains(preMsg, prefixMsg) && !strings.Contains(res, `(use "git commit" to conclude merge)`) {
					assert.Must(utils.RunShell(ctx, "git", "commit", "--amend", "--no-edit", "-m", strconv.Quote(msg)))
				} else {
					assert.Must(utils.RunShell(ctx, "git", "commit", "-m", strconv.Quote(msg)))
				}

				s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
					s.Prefix = "push git message: "
				})
				s.Start()
				pushOutput := assert.Must1(utils.RunOutput(ctx, "git", "push", "--force-with-lease", "origin", utils.GetBranchName()))
				if shouldPullDueToRemoteUpdate(pushOutput) {
					err := gitPull()
					if err != nil {
						if isMergeConflict() {
							handleMergeConflict()
						} else {
							os.Exit(1)
						}
					} else {
						informUserToAmendAndPush()
					}
				}
				s.Stop()
				return
			}

			assert.Must(utils.RunShell(ctx, "git", "add", "--update"))

			diff := assert.Must1(utils.GetStagedDiff(ctx))
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
			msg = strings.TrimSpace(tap.Text(ctx, tap.TextOptions{
				Message:      "git message(update or enter):",
				InitialValue: msg,
				DefaultValue: msg,
				Placeholder:  "update or enter",
			}))

			if msg == "" {
				return
			}

			assert.Must(utils.RunShell(ctx, "git", "commit", "-m", strconv.Quote(msg)))
			assert.Must(utils.RunShell(ctx, "git", "push", "origin", utils.GetBranchName()))
			if flags.showPrompt {
				fmt.Println("\n" + generatePrompt + "\n")
			}
			log.Info().Any("usage", resp.Usage).Msg("openai response usage")
			return
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	return app
}

func shouldPullDueToRemoteUpdate(msg string) bool {
	return strings.Contains(msg, "stale info") ||
		strings.Contains(msg, "rejected") ||
		strings.Contains(msg, "failed to push") ||
		strings.Contains(msg, "remote rejected")
}

// 执行 git pull（默认 merge 模式）
func gitPull() error {
	cmd := exec.Command("git", "pull", "--no-rebase")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// 检查是否存在未解决的合并冲突（U=unmerged）
func isMergeConflict() bool {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// 处理合并冲突：打开编辑器让用户解决
func handleMergeConflict() {
	fmt.Println("❌ Merge conflicts detected! Please resolve them.")

	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, _ := cmd.Output()
	files := strings.Split(strings.TrimSpace(string(output)), "\n")

	editor := getEditor()

	for _, file := range files {
		if file == "" {
			continue
		}
		fmt.Printf("📝 Conflict in file: %s\n", file)

		editCmd := exec.Command(editor, file)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		fmt.Printf("Opening editor '%s'...\n", editor)
		if err := editCmd.Run(); err != nil {
			log.Printf("Failed to edit %s: %v", file, err)
		}
	}

	// 提示用户完成后续操作
	informUserToAmendAndPush()
}

func getEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}

	if _, err := exec.LookPath("zed"); err == nil {
		return "zed -w"
	}

	if _, err := exec.LookPath("code"); err == nil {
		return "code -w"
	}

	if _, err := exec.LookPath("vim"); err == nil {
		return "vim"
	}

	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}
	return "vi"
}

// 提示用户如何继续
func informUserToAmendAndPush() {
	fmt.Println("\n----------------------------------------")
	fmt.Println("🛠️  Conflict resolved or pulled successfully.")
	fmt.Println("Now you can:")
	fmt.Println("   1. Review changes")
	fmt.Println("   2. Run 'git add .' to stage resolved files")
	fmt.Println("   3. Run 'git commit' (do NOT use --amend yet unless you want to absorb merge)")
	fmt.Println("   4. Then do:")
	fmt.Println("      git push --force-with-lease")
	fmt.Println("")
	fmt.Println("💡 Tip: 如果你想保持单个 commit，可以在 merge 后做交互式 rebase：")
	fmt.Println("    git reset HEAD~1   # 取消 merge commit")
	fmt.Println("    git add .")
	fmt.Println("    git commit --amend")
	fmt.Println("    git push --force-with-lease")
	fmt.Println("----------------------------------------")

	fmt.Println("\nPress Enter after you're done...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
