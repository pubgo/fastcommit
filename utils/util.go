package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/v2/result"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/tidwall/match"
	_ "github.com/tidwall/match"
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

func RunShell(ctx context.Context, args ...string) (err error) {
	defer result.RecoveryErr(&err)
	now := time.Now()
	res := result.Wrap(RunOutput(ctx, args...)).Must()

	if res != "" {
		log.Info().Str("dur", time.Since(now).String()).Msgf("shell result: \n%s\n", res)
	}

	return nil
}

func RunOutput(ctx context.Context, args ...string) (_ string, gErr error) {
	defer result.RecoveryErr(&gErr, func(err error) error {
		if exitErr, ok := errors.AsA[exec.ExitError](err); ok && exitErr.String() == "signal: interrupt" {
			os.Exit(1)
		}

		return err
	})

	sh := getShell()
	if sh != "" {
		args = []string{sh, "-c", fmt.Sprintf(`'%s'`, strings.Join(args, " "))}
	}

	cmdLine := strings.TrimSpace(strings.Join(args, " "))
	log.Info().Msgf("shell: %s", cmdLine)

	args = result.Wrap(shell.Fields(cmdLine, nil)).Must(func(e *zerolog.Event) {
		e.Str("shell", cmdLine)
	})
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout)

	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderr)

	err := Run(
		func() error {
			return cmd.Start()
		},
		func() error {
			return cmd.Wait()
		},
	)

	if err.IsErr() {
		return "", fmt.Errorf("%s\nerr: %w", stderr.String(), err.GetErr())
	}

	output := stdout.Bytes()
	return strings.TrimSpace(string(output)), nil
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

// Your branch and 'origin/fix/version' have diverged,
// and have 1 and 1 different commits each, respectively.
//
//	(use "git pull" if you want to integrate the remote branch with yours)
//
// nothing to commit, working tree clean

func PreGitPush(ctx context.Context) (err error) {
	defer result.RecoveryErr(&err)

	isDirty := IsDirty().Must()
	if isDirty {
		return
	}

	res := result.Wrap(RunOutput(ctx, "git", "status")).Must()
	needPush := strings.Contains(res, "Your branch is ahead of") && strings.Contains(res, "(use \"git push\" to publish your local commits)")
	if !needPush {
		needPush =
			match.Match(res, "*Your branch and '*' have diverged*") &&
				strings.Contains(result.Wrap(RunOutput(ctx, "git", "reflog", "-1")).Must(), "(amend)")
	}

	if !needPush {
		return
	}

	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
		s.Prefix = "push git message: "
	})
	s.Start()
	res = assert.Must1(RunOutput(ctx, "git", "push", "--force-with-lease", "origin", GetBranchName()))
	s.Stop()

	if res == "" {
		return
	}
	return errors.New(res)
}

var GetBranchName = sync.OnceValue(func() string { return GetCurrentBranch().Must() })

func LoadConfigAndBranch() {
	branchName := GetBranchName()
	log.Info().Msg("branch: " + branchName)
	log.Info().Msg("config: " + configs.GetConfigPath())
}

func Run(executors ...func() error) result.Error {
	for _, executor := range executors {
		if err := executor(); err != nil {
			return result.ErrOf(errors.WrapCaller(err, 1))
		}
	}
	return result.Error{}
}

func getShell() string {
	sh := "bash"
	_, err := exec.LookPath(sh)
	if err == nil {
		return sh
	}

	sh = "sh"
	_, err = exec.LookPath(sh)
	if err == nil {
		return sh
	}

	return ""
}

func IsStatusNeedPush(msg string) bool {
	var pattern = `
*Your branch is ahead of '*' by * commits.
  (use "git push" to publish your local commits)*
`

	return match.Match(msg, pattern)
}
