package logcmd

import (
	"context"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "log",
		Short: "Show commit log (--oneline -20)",
		Children: []*redant.Command{
			{
				Use:   "graph",
				Short: "Show graph commit log",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return utils.RunGit(ctx, logArgs(true)...)
				},
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return utils.RunGit(ctx, logArgs(false)...)
		},
	}
}

func logArgs(graph bool) []string {
	if graph {
		return []string{"log", "--graph", "--decorate", "--oneline", "-30"}
	}
	return []string{"log", "--oneline", "-20"}
}
