package selfcmd

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v71/github"
	"github.com/hashicorp/go-getter"
	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/funk/assert"
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

			client := github.NewClient(http.DefaultClient)
			r, _ := lo.Must2(client.Repositories.GetLatestRelease(context.Background(), "pubgo", "fastcommit"))

			var p = tea.NewProgram(initialModel(r))
			mm := assert.Must1(p.Run()).(model)

			var downloadURL = mm.selected.GetBrowserDownloadURL()

			downloadDir := filepath.Join(os.TempDir(), "fastcommit")
			pwd := lo.Must(os.Getwd())

			execFile := filepath.Base(os.Args[0])
			execFile = lo.Must(exec.LookPath(execFile))

			client1 := &getter.Client{
				Ctx:              ctx,
				Src:              downloadURL,
				Dst:              downloadDir,
				Pwd:              pwd,
				Mode:             getter.ClientModeDir,
				ProgressListener: defaultProgressBar,
			}
			lo.Must0(client1.Get())
			lo.Must0(os.Rename(downloadDir+"/fastcommit", execFile))

			return nil
		},
	}
}
