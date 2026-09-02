package remotecmd

import (
	"context"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "remote",
		Short: "List remotes",
		Children: []*redant.Command{
			{
				Use:   "list",
				Short: "List remotes",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return utils.RunGit(ctx, remoteArgs()...)
				},
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return utils.RunGit(ctx, remoteArgs()...)
		},
	}
}

func remoteArgs() []string {
	return []string{"remote", "-v"}
}
