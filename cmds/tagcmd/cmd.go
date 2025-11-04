package tagcmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/tap"

	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/fastcommit/utils/fzfutil"
)

func New() *cli.Command {
	var flags = new(struct {
		fastCommit bool
	})

	return &cli.Command{
		Name:  "tag",
		Usage: "gen tag and push origin",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list all tags",
				Action: func(ctx context.Context, command *cli.Command) error {
					utils.Spin("fetch git tag: ", func() (r result.Result[any]) {
						utils.GitFetchAll(ctx)
						return
					})

					var tagText = strings.TrimSpace(utils.ShellExecOutput(ctx, "git", "tag", "-n", "--sort=-committerdate").Unwrap())
					tag, err := fzfutil.SelectWithFzf(ctx, strings.NewReader(tagText))
					if err != nil {
						return err
					}

					fmt.Println(tag)
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "fast",
				Usage:       "quickly generate tag",
				Value:       flags.fastCommit,
				Destination: &flags.fastCommit,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			utils.LogConfigAndBranch()

			if flags.fastCommit {
				tags := utils.GetAllGitTags(ctx)

				sort.Slice(tags, func(i, j int) bool { return tags[i].GreaterThanOrEqual(tags[j]) })
				selectTags := lo.Map(tags, func(item *semver.Version, index int) tap.SelectOption[*semver.Version] {
					return tap.SelectOption[*semver.Version]{
						Value: item,
						Label: item.Original(),
					}
				})
				selectTags = lo.Chunk(selectTags, 10)[0]

				tagResult := tap.Select[*semver.Version](ctx, tap.SelectOptions[*semver.Version]{
					Message: "git tag(enter):",
					Options: selectTags,
				})

				if tagResult == nil {
					return nil
				}

				tagName := tap.Text(ctx, tap.TextOptions{
					Message:      "git tag(enter):",
					InitialValue: tagResult.Original(),
					DefaultValue: tagResult.Original(),
					Placeholder:  "enter git tag",
					Validate: func(s string) error {
						if !strings.HasPrefix(s, "v") {
							return fmt.Errorf("tag name must start with v")
						}

						_, err := semver.NewSemver(s)
						if err == nil {
							return nil
						}
						return fmt.Errorf("tag is invalid, tag=%s err=%w", s, err)
					},
				})

				if tagName == "" {
					return fmt.Errorf("tag name is empty")
				}

				fmt.Println(utils.GitPushTag(ctx, tagName))
				return nil
			}

			var p = tea.NewProgram(initialModel())
			m := assert.Must1(p.Run()).(model)
			selected := strings.TrimSpace(m.selected)
			if selected == "" {
				return nil
			}

			tags := utils.GetAllGitTags(ctx)

			var ver *semver.Version
			if pathutil.IsExist(".version") {
				vv := strings.TrimPrefix(string(lo.Must1(os.ReadFile(".version"))), "v")
				tags = lo.Filter(tags, func(item *semver.Version, index int) bool { return item.Core().String() == vv })
				if len(tags) == 0 {
					ver = lo.Must1(semver.NewSemver(fmt.Sprintf("%s-%s.1", lo.Must1(os.ReadFile(".version")), selected)))
				} else {
					ver = utils.GetNextTag(selected, tags)
				}
			} else {
				ver = utils.GetNextTag(selected, tags)
			}

			if selected == envRelease {
				vv := string(lo.Must1(os.ReadFile(".version")))
				ver = lo.Must1(semver.NewSemver(vv))
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

			output := utils.GitPushTag(ctx, tagName)
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
