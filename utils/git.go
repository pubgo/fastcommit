package utils

import (
	"fmt"
	"github.com/pubgo/funk/errors"
	"os/exec"
	"strings"
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
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()

	if err != nil {
		return "", &KnownError{Message: "The current directory must be a Git repository!"}
	}

	return strings.TrimSpace(string(output)), nil
}

// ExcludeFromDiff 生成 Git 排除路径的格式
func ExcludeFromDiff(path string) string {
	return fmt.Sprintf(":(exclude)%s", path)
}

// filesToExclude 是需要排除的文件列表
var filesToExclude = []string{
	//"package-lock.json",
	//"pnpm-lock.yaml",
}

// GetStagedDiff 获取暂存区的差异
func GetStagedDiff(excludeFiles []string) (map[string]interface{}, error) {
	diffCached := []string{"diff", "--cached", "--diff-algorithm=minimal"}

	// 获取暂存区文件的名称
	fmt.Println(append(diffCached, append([]string{"--name-only"}, append(filesToExclude, excludeFiles...)...)...))
	cmdFiles := exec.Command("git", append(diffCached, append([]string{"--name-only"}, append(filesToExclude, excludeFiles...)...)...)...)
	filesOutput, err := cmdFiles.Output()
	if err != nil {
		return nil, errors.WrapCaller(err)
	}

	files := strings.Split(strings.TrimSpace(string(filesOutput)), "\n")
	if len(files) == 0 || files[0] == "" {
		return nil, nil
	}

	// 获取暂存区的完整差异
	cmdDiff := exec.Command("git", append(diffCached, append(filesToExclude, excludeFiles...)...)...)
	diffOutput, err := cmdDiff.Output()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"files": files,
		"diff":  strings.TrimSpace(string(diffOutput)),
	}, nil
}

// GetDetectedMessage 生成检测到的文件数量的消息
func GetDetectedMessage(files []string) string {
	fileCount := len(files)
	pluralSuffix := ""
	if fileCount > 1 {
		pluralSuffix = "s"
	}
	return fmt.Sprintf("Detected %d staged file%s", fileCount, pluralSuffix)
}
