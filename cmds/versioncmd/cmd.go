package versioncmd

import (
	"context"
	"fmt"

	"github.com/pubgo/fastcommit/utils"
	v1 "github.com/pubgo/fastcommit/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/version"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   utils.UsageDesc("%s version info", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			fmt.Println("project:", version.Project())
			fmt.Println("version:", version.Version())
			fmt.Println("release:", v1.ReleaseVersion())
			fmt.Println("commit-id:", version.CommitID())
			fmt.Println("build-time:", version.BuildTime())
			fmt.Println("device-id:", running.DeviceID)
			return nil
		},
	}
}
