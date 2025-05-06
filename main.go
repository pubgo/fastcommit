package main

import (
	"context"
	"fmt"
	_ "github.com/fynelabs/selfupdate"
	_ "github.com/google/go-github/v71/github"
	"github.com/hashicorp/go-getter"
	_ "github.com/hashicorp/go-getter"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	//bootstrap.Main()

	//cli := github.NewClient(http.DefaultClient)
	//r, _ := lo.Must2(cli.Repositories.GetLatestRelease(context.Background(), "pubgo", "fastcommit"))
	//pretty.Println(r)
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
}
