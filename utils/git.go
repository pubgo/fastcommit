package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitfield/script"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logfields"
	"github.com/pubgo/funk/v2/result"
	"github.com/rs/zerolog"
)

// KnownError 是一个自定义错误类型
type KnownError struct {
	Message string
}

func (e *KnownError) Error() string {
	return e.Message
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
func GetStagedDiff(ctx context.Context, excludeFiles ...string) (r result.Result[*GetStagedDiffRsp]) {
	defer result.Recovery(&r)
	diffCached := []string{"git", "diff", "--cached", "--diff-algorithm=minimal"}

	// 获取暂存区文件的名称
	filesOutput := RunOutput(ctx, append(diffCached, append([]string{"--name-only"}, excludeFiles...)...)...).Must()

	files := strings.Split(strings.TrimSpace(filesOutput), "\n")
	if len(files) == 0 || files[0] == "" {
		return r.WithValue(new(GetStagedDiffRsp))
	}

	// 获取暂存区的完整差异
	diffOutput := RunOutput(ctx, append(diffCached, excludeFiles...)...).Must()

	return r.WithValue(&GetStagedDiffRsp{
		Files: files,
		Diff:  strings.TrimSpace(diffOutput),
	})
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

func GitPushTag(ctx context.Context, ver string) string {
	if ver == "" {
		return ""
	}

	log.Info().Msg("git push tag " + ver)
	assert.Must(RunShell(ctx, "git", "tag", ver))
	return RunOutput(ctx, "git", "push", "origin", ver).Must()
}

func GitFetchAll(ctx context.Context) {
	assert.Must(RunShell(ctx, "git", "fetch", "--prune", "--tags"))
}

func IsDirty() (r result.Result[bool]) {
	output := result.Wrap(script.Exec("git status --porcelain").String()).
		Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to run git")
		})

	return result.MapTo(output, func(output string) bool {
		return len(strings.TrimSpace(output)) > 0
	})
}

func GetCommitCount(branch string) (r result.Result[int]) {
	shell := fmt.Sprintf("git rev-list %s --count", branch)
	output := result.Wrap(script.Exec(shell).String()).Log(func(e *zerolog.Event) {
		e.Str(logfields.Msg, fmt.Sprintf("failed to run shell %q", shell))
	})

	return result.FlatMapTo(output, func(count string) result.Result[int] {
		count = strings.TrimSpace(count)
		return result.Wrap(strconv.Atoi(count)).Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to parse git output")
		})
	})
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

func GetRepositoryName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}

	repoPath := strings.TrimSpace(string(output))
	return filepath.Base(repoPath), nil
}

// IsGitRepository checks if the current directory is inside a git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func GetCurrentBranchV1() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func ListAllBranches() ([]string, error) {
	// First, fetch to ensure we have latest remote branches
	fetchCmd := exec.Command("git", "fetch", "--prune")
	if err := fetchCmd.Run(); err != nil {
		// Continue even if fetch fails
		fmt.Printf("Warning: failed to fetch latest branches: %v\n", err)
	}

	// Get all branches (local and remote)
	cmd := exec.Command("git", "branch", "-a", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "HEAD") {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

func BranchExists(branch string) (bool, error) {
	// Check if it's a local branch
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", branch)
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	// Check if it's a remote branch
	remoteRef := branch
	if !strings.HasPrefix(branch, "origin/") {
		remoteRef = "origin/" + branch
	}

	cmd = exec.Command("git", "rev-parse", "--verify", "--quiet", remoteRef)
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	return false, nil
}

func DeleteBranch(branch string) error {
	// Use -D flag to force delete even if not merged
	cmd := exec.Command("git", "branch", "-D", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %s", branch, string(output))
	}
	return nil
}

// HasUncommittedChanges checks if there are uncommitted changes in the current worktree
func HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// HasUnpushedCommits checks if there are unpushed commits in the current branch
func HasUnpushedCommits() (bool, error) {
	branch, err := GetCurrentBranchV1()
	if err != nil {
		return false, err
	}

	// Check if the branch has an upstream
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err := cmd.Run(); err != nil {
		// No upstream branch configured
		// Check if the branch is already merged to main/master
		// This handles the case where the branch was merged and remote was deleted
		merged, mergeErr := IsMergedToOrigin("main")
		if mergeErr == nil && merged {
			// Branch is merged, so no unpushed commits
			return false, nil
		}

		// If we can't determine merge status or branch is not merged,
		// assume there are unpushed commits for safety
		return true, nil
	}

	// Check if there are commits ahead of upstream
	cmd = exec.Command("git", "rev-list", "--count", branch+"@{upstream}.."+branch)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check unpushed commits: %w", err)
	}

	count := strings.TrimSpace(string(output))
	return count != "0", nil
}

// IsMergedToOrigin checks if the current branch is merged to origin
func IsMergedToOrigin(targetBranch string) (bool, error) {
	currentBranch, err := GetCurrentBranchV1()
	if err != nil {
		return false, err
	}

	// Fetch the latest state from origin
	cmd := exec.Command("git", "fetch", "origin", targetBranch)
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch origin: %w", err)
	}

	// Check if the current branch is merged into origin/targetBranch
	cmd = exec.Command("git", "branch", "-r", "--contains", currentBranch)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check merge status: %w", err)
	}

	branches := strings.Split(string(output), "\n")
	targetRef := fmt.Sprintf("origin/%s", targetBranch)

	for _, branch := range branches {
		if strings.TrimSpace(branch) == targetRef {
			return true, nil
		}
	}

	return false, nil
}

// WorktreeInfo represents information about a git worktree
type WorktreeInfo struct {
	Path       string
	Branch     string
	Commit     string
	IsDetached bool
	IsCurrent  bool
}

// DetermineWorktreeNames determines the branch name and directory suffix based on input
// If input contains a slash, it's treated as a full branch name
// Otherwise, "/impl" is appended to create the branch name
func DetermineWorktreeNames(input string) (branchName, dirSuffix string) {
	if strings.Contains(input, "/") {
		// Input is a full branch name
		branchName = input
		// Sanitize for directory name
		dirSuffix = SanitizeBranchNameForDirectory(input)
	} else {
		// Input is an issue number or simple identifier
		branchName = fmt.Sprintf("%s/impl", input)
		dirSuffix = input
	}
	return branchName, dirSuffix
}

// CreateWorktree creates a new git worktree
func CreateWorktree(issueNumberOrBranch, baseBranch string) (string, error) {
	if !IsGitRepository() {
		return "", fmt.Errorf("not in a git repository")
	}

	repoName, err := GetRepositoryName()
	if err != nil {
		return "", err
	}

	// Get repository root directory
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}
	repoRoot := strings.TrimSpace(string(output))

	// Determine branch name and directory suffix
	branchName, dirSuffix := DetermineWorktreeNames(issueNumberOrBranch)

	// Create worktree directory path relative to repository root
	worktreeDir := filepath.Join(repoRoot, "..", fmt.Sprintf("%s-%s", repoName, dirSuffix))

	// Create the worktree
	cmd = exec.Command("git", "worktree", "add", worktreeDir, "-b", branchName, baseBranch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(worktreeDir)
	if err != nil {
		return worktreeDir, nil
	}

	return absPath, nil
}

// RemoveWorktree removes a git worktree by issue number or branch name
func RemoveWorktree(issueNumberOrBranch string) error {
	if !IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	repoName, err := GetRepositoryName()
	if err != nil {
		return err
	}

	// Get repository root directory
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}
	repoRoot := strings.TrimSpace(string(output))

	// Determine directory suffix
	_, dirSuffix := DetermineWorktreeNames(issueNumberOrBranch)

	// Create worktree directory path relative to repository root
	worktreeDir := filepath.Join(repoRoot, "..", fmt.Sprintf("%s-%s", repoName, dirSuffix))
	return RemoveWorktreeByPath(worktreeDir)
}

// RemoveWorktreeByPath removes a git worktree by its path
func RemoveWorktreeByPath(worktreePath string) error {
	if !IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	// Remove the worktree
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	return nil
}

// ListWorktrees returns a list of all worktrees
func ListWorktrees() ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []WorktreeInfo
	lines := strings.Split(string(output), "\n")
	var current WorktreeInfo

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = WorktreeInfo{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix if present
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		} else if line == "detached" {
			current.IsDetached = true
		} else if line == "" && current.Path != "" {
			worktrees = append(worktrees, current)
			current = WorktreeInfo{}
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	// Mark current worktree
	cwd, err := os.Getwd()
	if err == nil {
		for i := range worktrees {
			if absPath, err := filepath.Abs(worktrees[i].Path); err == nil {
				if strings.HasPrefix(cwd, absPath) {
					worktrees[i].IsCurrent = true
					break
				}
			}
		}
	}

	return worktrees, nil
}

// GetWorktreeForIssue finds a worktree for a specific issue number or branch name
func GetWorktreeForIssue(issueNumberOrBranch string) (*WorktreeInfo, error) {
	repoName, err := GetRepositoryName()
	if err != nil {
		return nil, err
	}

	// Determine directory suffix
	_, dirSuffix := DetermineWorktreeNames(issueNumberOrBranch)

	targetPath := fmt.Sprintf("%s-%s", repoName, dirSuffix)

	worktrees, err := ListWorktrees()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if strings.Contains(wt.Path, targetPath) {
			return &wt, nil
		}
	}

	return nil, fmt.Errorf("worktree for %s not found", issueNumberOrBranch)
}

// CreateWorktreeFromBranch creates a new git worktree from an existing branch
func CreateWorktreeFromBranch(worktreePath, sourceBranch, targetBranch string) error {
	if !IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	// Check if source branch starts with origin/
	isRemoteBranch := strings.HasPrefix(sourceBranch, "origin/")

	var cmd *exec.Cmd
	if isRemoteBranch {
		// For remote branches, create a new local branch tracking the remote
		cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", targetBranch, sourceBranch)
	} else {
		// For local branches, just check it out
		cmd = exec.Command("git", "worktree", "add", worktreePath, sourceBranch)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	return nil
}

// RunCommand executes a command in the current directory
func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func SanitizeBranchNameForDirectory(branchName string) string {
	// Define characters that are problematic in directory names across different OS
	// Windows: \ / : * ? " < > |
	// Unix: mainly / (null character is also problematic but unlikely in branch names)

	// Replace common path separators
	sanitized := strings.ReplaceAll(branchName, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, "\\", "-")

	// Replace other problematic characters
	re := regexp.MustCompile(`[*?:<>"|]`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Replace multiple consecutive hyphens with a single hyphen
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Trim leading and trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// If the result is empty (very unlikely), use a default
	if sanitized == "" {
		sanitized = "branch"
	}

	return sanitized
}
