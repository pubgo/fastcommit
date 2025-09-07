package upgradecmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-version"
	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/tap"

	"github.com/pubgo/fastcommit/utils/githubclient"
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
					releases := assert.Must1(client.List(ctx))

					tt := tablewriter.NewWriter(os.Stdout)
					tt.Header([]string{"Name", "OS", "Arch", "Size", "Url"})

					for _, r := range releases {
						for _, a := range githubclient.GetAssets(r) {
							if a.IsChecksumFile() {
								continue
							}

							assert.Must(tt.Append([]string{
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
			defer recovery.Exit(func(err error) error {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				errors.Debug(err)
				return err
			})

			client := githubclient.NewPublicRelease("pubgo", "fastcommit")
			r := assert.Must1(client.List(ctx))

			assets := githubclient.GetAssetList(r)
			assets = lo.Filter(assets, func(item githubclient.Asset, index int) bool { return !item.IsChecksumFile() })
			sort.Slice(assets, func(i, j int) bool {
				return assert.Must1(version.NewSemver(assets[i].Name)).GreaterThan(lo.Must(version.NewSemver(assets[j].Name)))
			})

			if len(assets) > 20 {
				assets = assets[:20]
			}

			result2 := tap.Select[string](context.Background(), tap.SelectOptions[string]{
				Message: "Which frontend framework do you prefer?",
				Options: lo.Map(assets, func(item githubclient.Asset, index int) tap.SelectOption[string] {
					return tap.SelectOption[string]{
						Value: item.Name,
						Label: fmt.Sprintf("%s %s %s", item.Name, item.OS, item.Arch),
					}
				}),
			})
			fmt.Printf("\nYou chose: %s\n", result2)

			asset, ok := lo.Find(assets, func(item githubclient.Asset) bool { return item.Name == result2 })
			assert.If(!ok, "%s not found", result2)
			var downloadURL = asset.URL

			downloadDir := filepath.Join(os.TempDir(), "fastcommit")
			pwd := assert.Must1(os.Getwd())

			execFile := filepath.Base(os.Args[0])
			execFile = assert.Must1(exec.LookPath(execFile))

			log.Info().Func(func(e *zerolog.Event) {
				e.Str("download_dir", downloadDir)
				e.Str("pwd", pwd)
				e.Str("exec_file", execFile)
				e.Msgf("start download %s", downloadURL)
			})

			c := &getter.Client{
				Ctx:              ctx,
				Src:              downloadURL,
				Dst:              downloadDir,
				Pwd:              pwd,
				Mode:             getter.ClientModeDir,
				ProgressListener: defaultProgressBar,
			}
			assert.Must(c.Get())
			assert.Must(os.Rename(downloadDir+"/fastcommit", execFile))

			return nil
		},
	}
}
