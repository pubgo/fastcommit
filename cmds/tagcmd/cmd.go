package tagcmd

import (
	"context"

	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name: "tag",
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			ver := utils.GetNextTag("alpha")
			utils.GitTag(ver.String())
			return nil
		},
	}
}
