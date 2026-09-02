package addcmd

import (
	"context"
	"fmt"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "add <file|.>",
		Short: "Stage files",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			args, err := addArgs(i.Args...)
			if err != nil {
				return err
			}

			return utils.RunGit(ctx, args...)
		},
	}
}

func addArgs(files ...string) ([]string, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("usage: add <file|.>")
	}

	return append([]string{"add"}, files...), nil
}
