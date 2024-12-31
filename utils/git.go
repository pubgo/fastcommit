package utils

import (
	"fmt"
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

// assertGitRepo 检查当前目录是否是 Git 仓库
func assertGitRepo() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()

	if err != nil {
		return "", &KnownError{Message: "The current directory must be a Git repository!"}
	}

	return strings.TrimSpace(string(output)), nil
}

// excludeFromDiff 生成 Git 排除路径的格式
func excludeFromDiff(path string) string {
	return fmt.Sprintf(":(exclude)%s", path)
}

// filesToExclude 是需要排除的文件列表
var filesToExclude = []string{
	"package-lock.json",
	"pnpm-lock.yaml",
	"*.lock", // yarn.lock, Cargo.lock, Gemfile.lock, Pipfile.lock, etc.
}

// getStagedDiff 获取暂存区的差异
func getStagedDiff(excludeFiles []string) (map[string]interface{}, error) {
	diffCached := []string{"diff", "--cached", "--diff-algorithm=minimal"}

	// 获取暂存区文件的名称
	cmdFiles := exec.Command("git", append(diffCached, append([]string{"--name-only"}, append(filesToExclude, excludeFiles...)...)...)...)
	filesOutput, err := cmdFiles.Output()
	if err != nil {
		return nil, err
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

// getDetectedMessage 生成检测到的文件数量的消息
func getDetectedMessage(files []string) string {
	fileCount := len(files)
	pluralSuffix := ""
	if fileCount > 1 {
		pluralSuffix = "s"
	}
	return fmt.Sprintf("Detected %d staged file%s", fileCount, pluralSuffix)
}

func main() {
	// 示例用法
	repoPath, err := assertGitRepo()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Git repository root:", repoPath)

	diff, err := getStagedDiff(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	if diff != nil {
		fmt.Println("Staged files:", diff["files"])
		fmt.Println("Staged diff:", diff["diff"])
		fmt.Println(getDetectedMessage(diff["files"].([]string)))
	}
}
