package rebasecmd

import (
	"context"
	"fmt"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	var flags = new(struct {
		cont  bool
		abort bool
		skip  bool
	})

	return &redant.Command{
		Use:   "rebase <upstream>",
		Short: "Rebase current branch",
		Options: []redant.Option{
			{
				Flag:        "continue",
				Description: "Continue in-progress rebase",
				Value:       redant.BoolOf(&flags.cont),
			},
			{
				Flag:        "abort",
				Description: "Abort in-progress rebase",
				Value:       redant.BoolOf(&flags.abort),
			},
			{
				Flag:        "skip",
				Description: "Skip current patch in rebase",
				Value:       redant.BoolOf(&flags.skip),
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			upstream := ""
			if len(i.Args) > 0 {
				upstream = i.Args[0]
			}

			args, err := rebaseArgs(upstream, flags.cont, flags.abort, flags.skip)
			if err != nil {
				return err
			}

			return utils.RunGit(ctx, args...)
		},
	}
}

func rebaseArgs(upstream string, cont, abort, skip bool) ([]string, error) {
	flags := 0
	for _, on := range []bool{cont, abort, skip} {
		if on {
			flags++
		}
	}

	if flags > 1 {
		return nil, fmt.Errorf("usage: rebase takes only one of --continue/--abort/--skip")
	}

	if flags == 1 && upstream != "" {
		return nil, fmt.Errorf("usage: rebase <upstream> cannot be combined with --continue/--abort/--skip")
	}

	if cont {
		return []string{"rebase", "--continue"}, nil
	}
	if abort {
		return []string{"rebase", "--abort"}, nil
	}
	if skip {
		return []string{"rebase", "--skip"}, nil
	}

	if upstream == "" {
		return nil, fmt.Errorf("usage: rebase <upstream> or one of --continue/--abort/--skip")
	}

	return []string{"rebase", upstream}, nil
}
