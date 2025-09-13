package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
	"mvdan.cc/sh/v3/shell"
)

func GetAllRemoteTags(ctx context.Context) []*semver.Version {
	log.Info().Msg("get all remote tags")
	output := assert.Exit1(RunOutput(ctx, "git", "ls-remote", "--tags", "origin"))
	return lo.Map(strings.Split(output, "\n"), func(item string, index int) *semver.Version {
		item = strings.TrimSpace(item)
		if !strings.HasPrefix(item, "refs/tags/") {
			return nil
		}

		item = strings.TrimPrefix(item, "refs/tags/")
		if !strings.HasPrefix(item, "v") {
			return nil
		}

		vv, err := semver.NewSemver(item)
		if err != nil {
			log.Err(err).Str("tag", item).Msg("failed to parse git tag")
			assert.Must(err)
		}
		return vv
	})
}

func GetAllGitTags(ctx context.Context) []*semver.Version {
	log.Info().Msg("get all tags")
	var tagText = strings.TrimSpace(assert.Must1(RunOutput(ctx, "git", "tag")))
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

func GetCurMaxVer(ctx context.Context) *semver.Version {
	tags := GetAllGitTags(ctx)
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

func RunShell(ctx context.Context, args ...string) error {
	now := time.Now()
	res, err := RunOutput(ctx, args...)
	if err != nil {
		return errors.WrapCaller(err)
	}

	if res != "" {
		log.Info().Str("dur", time.Since(now).String()).Msgf("shell result: \n%s\n", res)
	}

	return nil
}

func RunOutput(ctx context.Context, args ...string) (_ string, gErr error) {
	defer result.Recovery(&gErr)

	var cmdLine = strings.Join(args, " ")
	log.Info().Msgf("shell: %s", strings.TrimSpace(cmdLine))

	args = result.Wrap(shell.Fields(cmdLine, nil)).Log().Must()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output := result.Wrap(cmd.Output()).
		Map(func(data []byte) []byte { return bytes.TrimSpace(data) }).
		Log().Must()
	return string(output), nil
}

func IsRemoteTagExist(err string) bool {
	return strings.Contains(err, "[rejected]") && strings.Contains(err, "tag already exists")
}

func IsRemotePushCommitFailed(err string) bool {
	return strings.Contains(err, "[rejected]") && strings.Contains(err, "failed to push some refs to")
}

func Spin[T any](name string, do func() result.Result[T]) result.Result[T] {
	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) { s.Prefix = name })
	s.Start()
	defer s.Stop()
	return do()
}
