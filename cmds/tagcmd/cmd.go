package tagcmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/fastcommit/utils/fzfutil"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "tag",
		Usage: "gen tag and push origin",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list all tags",
				Action: func(ctx context.Context, command *cli.Command) error {
					var tagText = strings.TrimSpace(utils.RunOutput(ctx, "git", "tag", "-n", "--sort=-committerdate").Must())
					tag, err := fzfutil.SelectWithFzf(ctx, strings.NewReader(tagText))
					if err != nil {
						return err
					}

					fmt.Println(tag)
					return nil
				},
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			utils.LoadConfigAndBranch()

			var p = tea.NewProgram(initialModel())
			m := assert.Must1(p.Run()).(model)
			selected := strings.TrimSpace(m.selected)
			if selected == "" {
				return nil
			}

			tags := utils.GetAllGitTags(ctx)
			ver := utils.GetNextTag(selected, tags)
			if selected == envRelease {
				ver = utils.GetNextReleaseTag(tags)
			}

			tagName := "v" + strings.TrimPrefix(ver.Original(), "v")
			var p1 = tea.NewProgram(InitialTextInputModel(tagName))
			m1 := assert.Must1(p1.Run()).(model2)
			if m1.exit {
				return nil
			}

			tagName = m1.Value()
			_, err := semver.NewVersion(tagName)
			if err != nil {
				return errors.Errorf("tag name is not valid: %s", tagName)
			}

			pushTag := result.Async(func() result.Result[string] {
				return result.OK(utils.GitPushTag(ctx, tagName))
			})
			time.Sleep(time.Millisecond * 10)

			output := utils.Spin("push tag: ", func() (r result.Result[string]) { return pushTag.Await() }).Must()
			if utils.IsRemoteTagExist(output) {
				utils.Spin("fetch git tag: ", func() (r result.Result[any]) {
					utils.GitFetchAll(ctx)
					return
				})
			}

			return nil
		},
	}
}
