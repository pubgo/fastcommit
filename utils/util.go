package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bitfield/script"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/typex"
	"github.com/samber/lo"
)

func GetGitTags() []*semver.Version {
	var tagText = strings.TrimSpace(assert.Exit1(RunOutput("git", "tag")))
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
	var tags = GetGitTags()
	var maxVer = GetGitMaxTag(tags)
	var preData = fmt.Sprintf("-%s.", pre)
	var curMaxVer = typex.DoBlock1(func() *semver.Version {
		preTags := lo.Filter(tags, func(item *semver.Version, index int) bool { return strings.Contains(item.String(), preData) })
		var curMaxVer = lo.MaxBy(preTags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
		return curMaxVer
	})

	var ver string
	if curMaxVer != nil && curMaxVer.GreaterThan(maxVer) {
		ver = strings.ReplaceAll(curMaxVer.Prerelease(), fmt.Sprintf("%s.", pre), "")
		ver = fmt.Sprintf("v%s-%s.%d", curMaxVer.Core().String(), pre, assert.Must1(strconv.Atoi(ver))+1)
	} else {
		ver = fmt.Sprintf("v%s-%s.1", maxVer.Core().String(), pre)
	}
	return assert.Exit1(semver.NewSemver(ver))
}

func GetGitMaxTag(tags []*semver.Version) *semver.Version {
	var maxVer = semver.Must(semver.NewVersion("v0.0.0"))

	for _, tag := range tags {
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

func RunShell(args ...string) error {
	var shell = strings.Join(args, " ")
	slog.Info(shell)
	return script.Exec(strings.Join(args, " ")).Error()
}

func RunOutput(args ...string) (string, error) {
	var shell = strings.Join(args, " ")
	slog.Info(shell)
	return script.Exec(strings.Join(args, " ")).String()
}
