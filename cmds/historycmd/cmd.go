package historycmd

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	mapset "github.com/deckarep/golang-set/v2"
	"os"
	"strconv"
	"strings"

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
			for _, line := range strings.Split(string(data), "\n") {
				lines := strings.SplitN(strings.TrimSpace(line), " ", 2)
				if len(lines) == 2 && isNumber(strings.Trim(strings.TrimSpace(lines[0]), "*")) {
					line = strings.TrimSpace(lines[1])
				}
				line = strings.TrimSpace(line)
				set.Add(line)
			}

			p := tea.NewProgram(initialModel(set.ToSlice()))
			_ = lo.Must(p.Run())
			return nil
		},
	}
}

func isNumber(str string) bool {
	if str == "" {
		return false
	}

	_, err := strconv.Atoi(str)
	return err == nil
}
