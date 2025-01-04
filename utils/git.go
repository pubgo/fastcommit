package utils

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
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
	return fmt.Sprintf("Detected %d staged file%s\n%s", fileCount, pluralSuffix, strings.Join(files, "\n"))
}

func GitPushTag(ver string) {
	assert.Exit(RunShell("git", "tag", ver))
	assert.Exit(RunShell("git", "push", "origin", ver))
}
