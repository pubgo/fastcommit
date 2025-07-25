package tagcmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/fastcommit/cmds/cmdutils"
	"github.com/pubgo/fastcommit/utils"
)

func New() *cli.Command {
	var pushRelease bool
	return &cli.Command{
		Name:  "tag",
		Usage: "gen tag and push origin",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "push",
				Usage:       "release tag and push remote",
				Value:       pushRelease,
				Destination: &pushRelease,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			cmdutils.LoadConfigAndBranch()

			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
				s.Prefix = "fetch git tag: "
			})
			s.Start()
			utils.GitFetchAll()
			s.Stop()

			var p = tea.NewProgram(initialModel())
			m := assert.Must1(p.Run()).(model)
			selected := strings.TrimSpace(m.selected)
			if selected == "" {
				return nil
			}

			tags := utils.GetAllGitTags()
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

			lo.Must0(os.MkdirAll("version", 0755))
			lo.Must0(os.WriteFile("version/.version", []byte(tagName), 0644))

			if pushRelease {
				utils.GitPushTag(tagName)
			}

			return nil
		},
	}
}
