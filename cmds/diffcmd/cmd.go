package diffcmd

import (
	"context"
	"fmt"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	var flags = new(struct {
		staged   bool
		unstaged bool
	})

	return &redant.Command{
		Use:   "diff",
		Short: "Show diff (--staged for staged, --unstaged for unstaged, default HEAD)",
		Options: []redant.Option{
			{
				Flag:        "staged",
				Description: "Show staged diff",
				Value:       redant.BoolOf(&flags.staged),
			},
			{
				Flag:        "unstaged",
				Description: "Show unstaged diff",
				Value:       redant.BoolOf(&flags.unstaged),
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			args, err := diffArgs(flags.staged, flags.unstaged)
			if err != nil {
				return err
			}

			return utils.RunGit(ctx, args...)
		},
	}
}

func diffArgs(staged, unstaged bool) ([]string, error) {
	if staged && unstaged {
		return nil, fmt.Errorf("usage: diff takes only one of --staged/--unstaged")
	}

	if staged {
		return []string{"diff", "--cached"}, nil
	}
	if unstaged {
		return []string{"diff"}, nil
	}

	return []string{"diff", "HEAD"}, nil
}
