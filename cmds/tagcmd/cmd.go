package tagcmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name: "tag",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			var p = tea.NewProgram(initialModel())
			m := assert.Must1(p.Run()).(model)
			var tags = utils.GetGitTags()
			ver := utils.GetNextTag(m.selected, tags)
			if m.selected == "release" {
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
				return errors.Format("tag name is not valid: %s", tagName)
			}
			slog.Info(fmt.Sprintf("selected tag: %s", tagName))
			utils.GitPushTag(tagName)
			return nil
		},
	}
}
