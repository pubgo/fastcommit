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
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/typex"
	"github.com/samber/lo"
)

func GetAllGitTags() []*semver.Version {
	var tagText = strings.TrimSpace(assert.Exit1(RunOutput("git", "tag")))
	var tags = strings.Split(tagText, "\n")
	var versions = make([]*semver.Version, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		vv, err := semver.NewSemver(tag)
		if err != nil {
			slog.Error("failed to parse git tag", "tag", tag, "err", err)
			assert.Exit(err)
		}
		versions = append(versions, vv)
	}
	return versions
}

func GetNextReleaseTag(tags []*semver.Version) *semver.Version {
	var curMaxVer = typex.DoBlock1(func() *semver.Version {
		return lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
	})

	if curMaxVer.Prerelease() == "" {
		segments := curMaxVer.Core().Segments()
		return assert.Exit1(semver.NewSemver(fmt.Sprintf("v%d.%d.%d", segments[0], segments[1], segments[2]+1)))
	}

	return curMaxVer.Core()
}

func GetNextTag(pre string, tags []*semver.Version) *semver.Version {
	var maxVer = GetGitMaxTag(tags)
	var curMaxVer = typex.DoBlock1(func() *semver.Version {
		return lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
	})

	var ver string
	if curMaxVer != nil && curMaxVer.Core().GreaterThan(maxVer) {
		ver = strings.ReplaceAll(curMaxVer.Prerelease(), fmt.Sprintf("%s.", pre), "")
		if ver == "" {
			ver = "1"
		}

		ver = fmt.Sprintf("v%s-%s.%d", curMaxVer.Core().String(), pre, assert.Must1(strconv.Atoi(ver))+1)
	} else {
		ver = fmt.Sprintf("v%s-%s.1", maxVer.String(), pre)
	}
	return assert.Exit1(semver.NewSemver(ver))
}

func GetGitMaxTag(tags []*semver.Version) *semver.Version {
	if len(tags) == 0 {
		return semver.Must(semver.NewVersion("v0.0.1"))
	}

	maxVer := semver.Must(semver.NewVersion("v0.0.0"))
	for _, tag := range tags {
		if maxVer.Compare(tag) >= 0 {
			continue
		}

		maxVer = tag
	}

	segments := maxVer.Segments()
	v3 := segments[2]
	v3 = lo.If(strings.Contains(maxVer.String(), "-"), v3).Else(v3 + 1)

	return semver.Must(semver.NewVersion(fmt.Sprintf("v%d.%d.%d", segments[0], segments[1], v3)))
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
	result, err := RunOutput(args...)
	if err != nil {
		return errors.WrapCaller(err)
	}
	result = strings.TrimSpace(result)
	if result != "" {
		slog.Info(result)
	}

	return nil
}

func RunOutput(args ...string) (string, error) {
	var shell = strings.Join(args, " ")
	slog.Info(shell)
	return script.Exec(shell).String()
}
