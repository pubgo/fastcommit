package statuscmd

import (
	"context"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "status",
		Short: "Show working tree status",
		Children: []*redant.Command{
			{
				Use:   "short",
				Short: "Show concise status",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return utils.RunGit(ctx, statusArgs(true)...)
				},
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return utils.RunGit(ctx, statusArgs(false)...)
		},
	}
}

func statusArgs(short bool) []string {
	if short {
		return []string{"status", "--short"}
	}
	return []string{"status"}
}
