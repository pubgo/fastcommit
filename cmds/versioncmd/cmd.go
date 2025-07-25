package versioncmd

import (
	"context"
	"fmt"

	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: utils.UsageDesc("%s version info", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			ver := version.Version()
			fmt.Println("project:", version.Project())
			fmt.Println("version:", ver)
			fmt.Println("commit-id:", version.CommitID())
			fmt.Println("build-time:", version.BuildTime())
			fmt.Println("instance-id:", running.InstanceID)
			return nil
		},
	}
}
