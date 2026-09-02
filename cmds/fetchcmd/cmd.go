package fetchcmd

import (
	"context"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	var flags = new(struct {
		prune bool
	})

	return &redant.Command{
		Use:   "fetch",
		Short: "Fetch from remote",
		Options: []redant.Option{
			{
				Flag:        "prune",
				Description: "Fetch and prune stale refs",
				Value:       redant.BoolOf(&flags.prune),
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return utils.RunGit(ctx, fetchArgs(flags.prune)...)
		},
	}
}

func fetchArgs(prune bool) []string {
	if prune {
		return []string{"fetch", "--prune"}
	}
	return []string{"fetch"}
}
