package upgradecmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/go-getter"
	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/fastcommit/utils/githubclient"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "self upgrade management",
		Commands: []*cli.Command{
			{
				Name: "list",
				Action: func(ctx context.Context, command *cli.Command) error {
					client := githubclient.NewPublicRelease("pubgo", "fastcommit")
					releases := lo.Must(client.List(ctx))

					tt := tablewriter.NewWriter(os.Stdout)
					tt.Header([]string{"Name", "OS", "Arch", "Size", "Url"})

					for _, r := range releases {
						for _, a := range githubclient.GetAssets(r) {
							if a.IsChecksumFile() {
								continue
							}

							lo.Must0(tt.Append([]string{
								a.Name,
								a.OS,
								a.Arch,
								githubclient.GetSizeFormat(a.Size),
								a.URL,
							}))
						}
					}
					return tt.Render()
				},
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			client := githubclient.NewPublicRelease("pubgo", "fastcommit")
			r := lo.Must(client.List(ctx))

			var p = tea.NewProgram(initialModel(githubclient.GetAssetList(r)))
			mm := assert.Must1(p.Run()).(*model)

			var downloadURL = mm.selected.URL

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
