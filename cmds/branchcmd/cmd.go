package branchcmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pubgo/fastgit/utils"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	var flags = new(struct {
		listRemote bool
	})

	return &redant.Command{
		Use:   "branch",
		Short: "Branch inspect and switch shortcuts",
		Children: []*redant.Command{
			{
				Use:   "current",
				Short: "Show current branch",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return utils.RunGit(ctx, "branch", "--show-current")
				},
			},
			{
				Use:   "list",
				Short: "List branches (--remote for remote branches)",
				Options: []redant.Option{
					{
						Flag:        "remote",
						Description: "List remote branches",
						Value:       redant.BoolOf(&flags.listRemote),
					},
				},
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					if flags.listRemote {
						return utils.RunGit(ctx, "branch", "-r")
					}
					return utils.RunGit(ctx, "branch")
				},
			},
			{
				Use:   "checkout",
				Short: "Checkout branch",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					name, err := requireOneArg(i.Args, "branch checkout <name>")
					if err != nil {
						return err
					}
					return utils.RunGit(ctx, "checkout", name)
				},
			},
			{
				Use:   "checkout-remote",
				Short: "Checkout remote branch to a tracking local branch",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					name, err := requireOneArg(i.Args, "branch checkout-remote <name>")
					if err != nil {
						return err
					}
					args, err := checkoutRemoteArgs(name)
					if err != nil {
						return err
					}
					return utils.RunGit(ctx, args...)
				},
			},
			{
				Use:   "create",
				Short: "Create and checkout new branch",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					name, err := requireOneArg(i.Args, "branch create <name>")
					if err != nil {
						return err
					}
					return utils.RunGit(ctx, "checkout", "-b", name)
				},
			},
			{
				Use:   "delete",
				Short: "Delete local branch",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					name, err := requireOneArg(i.Args, "branch delete <name>")
					if err != nil {
						return err
					}
					return utils.RunGit(ctx, "branch", "-d", name)
				},
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return redant.DefaultHelpFn()(ctx, i)
		},
	}
}

func requireOneArg(args []string, usage string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("usage: %s", usage)
	}
	return strings.TrimSpace(args[0]), nil
}

func checkoutRemoteArgs(name string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("usage: branch checkout-remote <name>")
	}

	remote := name
	if !strings.HasPrefix(remote, "origin/") {
		remote = "origin/" + remote
	}

	local := strings.TrimPrefix(remote, "origin/")
	return []string{"checkout", "-b", local, "--track", remote}, nil
}
