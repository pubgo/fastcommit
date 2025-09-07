package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bitfield/script"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/typex"
	"github.com/samber/lo"
)

func GetAllGitTags() []*semver.Version {
	log.Info().Msg("get all tags")
	var tagText = strings.TrimSpace(assert.Must1(RunOutput("git", "tag")))
	var tags = strings.Split(tagText, "\n")
	var versions = make([]*semver.Version, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if !strings.HasPrefix(tag, "v") {
			continue
		}

		vv, err := semver.NewSemver(tag)
		if err != nil {
			log.Err(err).Str("tag", tag).Msg("failed to parse git tag")
			assert.Must(err)
		}
		versions = append(versions, vv)
	}
	return versions
}

func GetCurMaxVer() *semver.Version {
	tags := GetAllGitTags()
	return typex.DoBlock1(func() *semver.Version {
		return lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
	})
}

func GetNextReleaseTag(tags []*semver.Version) *semver.Version {
	var curMaxVer = typex.DoBlock1(func() *semver.Version {
		return lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
	})

	if curMaxVer.Prerelease() == "" {
		segments := curMaxVer.Core().Segments()
		return assert.Must1(semver.NewSemver(fmt.Sprintf("v%d.%d.%d", segments[0], segments[1], segments[2]+1)))
	}

	return curMaxVer.Core()
}

func GetNextTag(pre string, tags []*semver.Version) *semver.Version {
	var maxVer = GetNextGitMaxTag(tags)
	var curMaxVer = typex.DoBlock1(func() *semver.Version {
		tags = lo.Filter(tags, func(item *semver.Version, index int) bool { return strings.Contains(item.String(), pre) })
		return lo.MaxBy(tags, func(a *semver.Version, b *semver.Version) bool { return a.Compare(b) > 0 })
	})

	var ver string
	if curMaxVer != nil && curMaxVer.Core().GreaterThanOrEqual(maxVer) {
		ver = strings.ReplaceAll(curMaxVer.Prerelease(), fmt.Sprintf("%s.", pre), "")
		ver = fmt.Sprintf("v%s-%s.%d", curMaxVer.Core().String(), pre, assert.Must1(strconv.Atoi(ver))+1)
	} else {
		ver = fmt.Sprintf("v%s-%s.1", maxVer.Core().String(), pre)
	}
	return assert.Must1(semver.NewSemver(ver))
}

func GetNextGitMaxTag(tags []*semver.Version) *semver.Version {
	maxVer := semver.Must(semver.NewVersion("v0.0.1"))
	if len(tags) == 0 {
		return maxVer
	}

	for _, tag := range tags {
		if maxVer.Compare(tag) >= 0 {
			continue
		}

		maxVer = tag
	}

	segments := maxVer.Segments()
	v3Segment := lo.If(strings.Contains(maxVer.String(), "-"), segments[2]).Else(segments[2] + 1)

	return semver.Must(semver.NewVersion(fmt.Sprintf("v%d.%d.%d", segments[0], segments[1], v3Segment)))
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
		log.Info().Msg("shell result")
		fmt.Println(result)
	}

	return nil
}

func RunOutput(args ...string) (string, error) {
	var shell = strings.Join(args, " ")
	log.Info().Msg("shell: " + strings.TrimSpace(shell))
	return script.Exec(shell).String()
}
