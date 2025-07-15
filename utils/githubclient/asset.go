package githubclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/google/go-github/v71/github"
)

func getAssets(repo *github.RepositoryRelease) Assets {
	var assetList Assets
	for _, a := range repo.Assets {
		assetList = append(assetList, Asset{
			Name:      repo.GetTagName(),
			URL:       a.GetBrowserDownloadURL(),
			Type:      a.GetContentType(),
			Size:      a.GetSize(),
			CreatedAt: a.GetCreatedAt().Time,
			OS:        getOS(a.GetName()),
			Arch:      getArch(a.GetName()),
		})
	}
	return assetList
}

type Asset struct {
	Name, OS, Arch, URL, Type string
	Size                      int
	CreatedAt                 time.Time
}

func (a Asset) IsChecksumFile() bool {
	// maximum file size 64KB
	return checksumRe.MatchString(strings.ToLower(a.Name)) && a.Size < 64*1024
}

func (a Asset) Key() string {
	return a.OS + "/" + a.Arch
}

func (a Asset) Is32Bit() bool {
	return a.Arch == "386"
}

func (a Asset) IsMac() bool {
	return a.OS == "darwin"
}
func (a Asset) IsWindows() bool {
	return a.OS == "windows"
}

func (a Asset) IsLinux() bool {
	return a.OS == "linux"
}

func (a Asset) IsMacM1() bool {
	return a.IsMac() && a.Arch == "arm64"
}

type Assets []Asset

func (as Assets) HasM1() bool {
	//detect if we have a native m1 asset
	for _, a := range as {
		if a.IsMacM1() {
			return true
		}
	}
	return false
}

func checkExt(url string, size int, name string) error {
	fext := getFileExt(url)
	if fext == "" && size > 1024*1024 {
		fext = ".bin" // +1MB binary
	}

	switch fext {
	case ".bin", ".zip", ".tar.bz", ".tar.bz2", ".bz2", ".gz", ".tar.gz", ".tgz":
		// valid
		return nil
	default:
		return fmt.Errorf("fetched asset has unsupported file type: %s (ext '%s')", name, fext)
	}
}

func getSizeFormat(size int) string {
	return units.HumanSize(float64(size))
}
