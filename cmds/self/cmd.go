package selfcmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v71/github"
	"github.com/hashicorp/go-getter"
	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "self",
		Usage: "self upgrade management",
		Commands: []*cli.Command{
			{
				Name: "list",
				Action: func(ctx context.Context, command *cli.Command) error {
					client := github.NewClient(http.DefaultClient)
					releaseList, _ := lo.Must2(client.Repositories.ListReleases(ctx, "pubgo", "fastcommit", nil))

					tt := tablewriter.NewWriter(os.Stdout)
					tt.SetHeader([]string{"Tag", "Name", "Url"})
					tt.SetBorder(true)
					tt.SetRowLine(true)

					for _, r := range releaseList {
						for _, a := range r.Assets {
							tt.Append([]string{lo.FromPtr(r.TagName), lo.FromPtr(a.Name), lo.FromPtr(a.BrowserDownloadURL)})
						}
					}
					tt.Render()
					return nil
				},
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			cli := github.NewClient(http.DefaultClient)
			r, _ := lo.Must2(cli.Repositories.GetLatestRelease(context.Background(), "pubgo", "fastcommit"))
			pretty.Println(r)
			return nil
			var sss = "https://github.com/pubgo/fastcommit/releases/download/v0.0.6-alpha.6/fastcommit_Darwin_x86_64.tar.gz"
			var opts []getter.ClientOption

			//ff := lo.Must(os.CreateTemp("fastcommit", "v0.0.6-alpha.6"))

			fffff := filepath.Join(os.TempDir(), "fastcommit")
			fmt.Println(fffff)
			pwd := lo.Must(os.Getwd())

			fmt.Println(runtime.GOOS)
			fmt.Println(runtime.GOARCH)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Build the client
			client := &getter.Client{
				Ctx:              ctx,
				Src:              sss,
				Dst:              fffff,
				Pwd:              pwd,
				Mode:             getter.ClientModeDir,
				Options:          opts,
				ProgressListener: defaultProgressBar,
			}
			lo.Must0(client.Get())

			return nil
		},
	}
}
