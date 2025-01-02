package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/samber/lo"
)

func GetGitTags() []*semver.Version {
	var tagText = strings.TrimSpace(string(assert.Exit1(ShellOutput("git", "tag"))))
	var tags = strings.Split(tagText, "\n")
	var versions = make([]*semver.Version, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		vv, err := semver.NewSemver(tag)
		if err != nil {
			continue
		}
		versions = append(versions, vv)
	}
	return versions
}

func GetNextTag(pre string) *semver.Version {
	var maxVer = GetGitMaxTag()
	var preData = fmt.Sprintf("-%s.", pre)
	var tags = lo.Filter(GetGitTags(), func(item *semver.Version, index int) bool { return strings.Contains(item.String(), preData) })
	var curMaxVer = lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })

	var ver string
	if curMaxVer != nil && curMaxVer.GreaterThan(maxVer) {
		ver = strings.ReplaceAll(curMaxVer.Prerelease(), fmt.Sprintf("%s.", pre), "")
		ver = fmt.Sprintf("v%s-%s.%d", curMaxVer.Core().String(), pre, assert.Must1(strconv.Atoi(ver))+1)
	} else {
		ver = fmt.Sprintf("v%s-alpha.1", maxVer.Core().String())
	}
	return assert.Exit1(semver.NewSemver(ver))
}

func GetGitMaxTag() *semver.Version {
	var maxVer = semver.Must(semver.NewVersion("v0.0.0"))

	for _, tag := range GetGitTags() {
		if strings.Contains(tag.String(), "-") {
			continue
		}

		if maxVer.Compare(tag) >= 0 {
			continue
		}

		maxVer = tag
	}

	return maxVer
}

func UsageDesc(format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args...)
	return strings.ToUpper(s[0:1]) + s[1:]
}

func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
			cancel()
		}
	}()
	return ctx
}

func IsHelp() bool {
	help := strings.TrimSpace(os.Args[len(os.Args)-1])
	if strings.HasSuffix(help, "--help") || strings.HasSuffix(help, "-h") {
		return true
	}
	return false
}
