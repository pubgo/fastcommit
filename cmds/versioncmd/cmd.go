package versioncmd

import (
	"context"
	"fmt"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/fastcommit/version"
	"github.com/pubgo/funk/v2/buildinfo"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   utils.UsageDesc("%s version info", buildinfo.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			fmt.Println("project:", buildinfo.Project())
			fmt.Println("version:", buildinfo.Version())
			fmt.Println("release:", version.ReleaseVersion())
			fmt.Println("commit-id:", buildinfo.CommitID())
			fmt.Println("build-time:", buildinfo.BuildTime())
			fmt.Println("device-id:", running.DeviceID)
			return nil
		},
	}
}
