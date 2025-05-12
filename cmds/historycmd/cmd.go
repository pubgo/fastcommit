package historycmd

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "history",
		Usage: "shell history command management",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			var path = "/Users/barry/Documents/git/zshrc.history"
			var data = lo.Must(os.ReadFile(path))
			var set = mapset.NewSet[string]()
			_ = set
			for _, line := range strings.Split(string(data), "\n") {
				lines := strings.SplitN(strings.TrimSpace(line), " ", 2)
				if len(lines) == 2 {
					line = lines[1]
				} else {
					line = lines[0]
				}
				line = strings.TrimSpace(line)
				set.Add(line)
			}

			for _, v := range set.ToSlice() {
				fmt.Println(v)
			}
			return nil

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
				return errors.Format("tag name is not valid: %s", tagName)
			}

			utils.GitPushTag(tagName)
			return nil
		},
	}
}
