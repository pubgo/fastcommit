package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bitfield/script"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
)

// KnownError 是一个自定义错误类型
type KnownError struct {
	Message string
}

func (e *KnownError) Error() string {
	return e.Message
}

// AssertGitRepo 检查当前目录是否是 Git 仓库
func AssertGitRepo() (string, error) {
	output, err := RunOutput("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", &KnownError{Message: "The current directory must be a Git repository!"}
	}

	return strings.TrimSpace(output), nil
}

// ExcludeFromDiff 生成 Git 排除路径的格式
func ExcludeFromDiff(path string) string {
	return fmt.Sprintf(":(exclude)%s", path)
}

type GetStagedDiffRsp struct {
	Files []string `json:"files"`
	Diff  string   `json:"diff"`
}

// GetStagedDiff 获取暂存区的差异
func GetStagedDiff(excludeFiles []string) (*GetStagedDiffRsp, error) {
	diffCached := []string{"git", "diff", "--cached", "--diff-algorithm=minimal"}

	// 获取暂存区文件的名称
	filesOutput, err := RunOutput(append(diffCached, append([]string{"--name-only"}, excludeFiles...)...)...)
	if err != nil {
		return nil, errors.WrapCaller(err)
	}

	files := strings.Split(strings.TrimSpace(filesOutput), "\n")
	if len(files) == 0 || files[0] == "" {
		return new(GetStagedDiffRsp), nil
	}

	// 获取暂存区的完整差异
	diffOutput, err := RunOutput(append(diffCached, excludeFiles...)...)
	if err != nil {
		return nil, errors.WrapCaller(err)
	}

	return &GetStagedDiffRsp{
		Files: files,
		Diff:  strings.TrimSpace(diffOutput),
	}, nil
}

// GetDetectedMessage 生成检测到的文件数量的消息
func GetDetectedMessage(files []string) string {
	fileCount := len(files)
	pluralSuffix := ""
	if fileCount > 1 {
		pluralSuffix = "s"
	}
	return fmt.Sprintf("detected %d staged file%s", fileCount, pluralSuffix)
}

func GitPushTag(ver string) string {
	if ver == "" {
		return ""
	}

	log.Info().Msg("git push tag " + ver)
	assert.Must(RunShell("git", "tag", ver))
	return assert.Must1(RunOutput("git", "push", "origin", ver))
}

func GitFetchAll() {
	assert.Must(RunShell("git", "fetch", "--prune", "--tags"))
}

func IsDirty() (r result.Result[bool]) {
	output := result.Wrap(script.Exec("git status --porcelain").String()).
		MapErr(func(err error) error {
			return fmt.Errorf("failed to run git: %w", err)
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	return r.WithValue(len(strings.TrimSpace(string(output))) > 0)
}

func GetCommitCount(branch string) (r result.Result[int]) {
	shell := fmt.Sprintf("git rev-list %s --count", branch)
	output := result.Wrap(script.Exec(shell).String()).
		MapErr(func(err error) error {
			return fmt.Errorf("failed to run shell %q, err=%w", shell, err)
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	count := result.Wrap(strconv.Atoi(strings.TrimSpace(output))).
		MapErr(func(err error) error {
			return fmt.Errorf("failed to parse git output: err=%w", err)
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	return r.WithValue(count)
}

func GetCurrentBranch() result.Result[string] {
	shell := "git branch --show-current"
	return result.Wrap(script.Exec(shell).String()).
		Map(func(s string) string {
			return strings.TrimSpace(s)
		}).
		MapErr(func(err error) error {
			return fmt.Errorf("failed to run shell %q, err=%w", shell, err)
		})
}

func PushTag(tag string) result.Error {
	shell := fmt.Sprintf("git push origin %s", tag)
	return result.ErrOf(script.Exec(shell).Error()).Map(func(err error) error {
		return fmt.Errorf("failed to run shell %q, err=%w", shell, err)
	})
}
