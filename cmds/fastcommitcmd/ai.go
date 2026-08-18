package fastcommitcmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixcontext"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/yarlson/tap"

	"github.com/pubgo/fastgit/pkg/aiprovider"
	"github.com/pubgo/fastgit/pkg/gitconflict"
	"github.com/pubgo/fastgit/pkg/repoconfig"
	"github.com/pubgo/fastgit/pkg/workflow"
	"github.com/pubgo/fastgit/utils"
)

func runAICommit(ctx context.Context, flags *flagOptions) error {
	di := dixcontext.Get(ctx)
	var params cmdParams
	params = dix.Inject(di, params)

	utils.LogConfigAndBranch()

	if flags.fastCommit {
		return runFastCommit(ctx, flags)
	}
	return runNormalCommit(ctx, flags, params)
}

func runFastCommit(ctx context.Context, flags *flagOptions) error {
	isDirty := utils.IsDirty().Unwrap()
	if !isDirty {
		return nil
	}

	preMsg := strings.TrimSpace(utils.ShellExecOutput(ctx, "git", "log", "-1", "--pretty=%B").Unwrap())
	prefixMsg := fmt.Sprintf("chore: quick update %s", utils.GetBranchName())
	msg := fmt.Sprintf("%s at %s", prefixMsg, time.Now().Format(time.DateTime))
	if flags.edit {
		msg = strings.TrimSpace(tap.Text(ctx, tap.TextOptions{
			Message:      "git message(update or enter):",
			InitialValue: msg,
			DefaultValue: msg,
			Placeholder:  "update or enter",
		}))
		if msg == "" {
			return nil
		}
	}

	repoRoot := mustRepoRoot()
	repoCfg, _ := repoconfig.Load(repoRoot)
	if err := enforceRepoPolicy(repoCfg, currentBranch(), msg, flags.skipPolicy); err != nil {
		return err
	}
	warnRepoPolicy(repoCfg, currentBranch(), msg)

	assert.Must(utils.ShellExec(ctx, "git", "add", "-A"))
	status := utils.ShellExecOutput(ctx, "git", "status").Unwrap()

	if err := runPreCommitCheck(ctx, repoRoot, flags.skipCheck, true); err != nil {
		return err
	}

	if flags.amend && strings.Contains(preMsg, prefixMsg) && !strings.Contains(status, `(use "git commit" to conclude merge)`) {
		if err := utils.GitCommit(ctx, msg, "--amend"); err != nil {
			return err
		}
	} else {
		if err := utils.GitCommit(ctx, msg); err != nil {
			return err
		}
	}

	if err := ensurePushPolicy(repoRoot, utils.GetBranchName(), flags.overridePolicy); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "→ pushing to remote...")
	return finishPush(ctx)
}

func runNormalCommit(ctx context.Context, flags *flagOptions, params cmdParams) error {
	// Stage first, check, then AI — soft-reset squash happens only after checks succeed.
	fmt.Fprintln(os.Stderr, "→ staging changes...")
	if utils.IsDirty().Unwrap() {
		assert.Must(utils.ShellExec(ctx, "git", "add", "-A"))
	}

	diffResult := utils.GetStagedDiff(ctx).Unwrap()
	if diffResult == nil || len(diffResult.Files) == 0 {
		fmt.Fprintln(os.Stderr, "→ nothing to commit")
		return nil
	}

	repoRoot := mustRepoRoot()
	if err := runPreCommitCheck(ctx, repoRoot, flags.skipCheck, false); err != nil {
		return err
	}

	repoCfg, _ := repoconfig.Load(repoRoot)
	if err := repoCfg.CheckBranch(currentBranch(), flags.skipPolicy); err != nil {
		return err
	}
	for _, file := range diffResult.Files {
		if repoCfg.MatchesSensitivePath(file) {
			log.Warn().Str("file", file).Msg("sensitive path staged — review carefully")
		}
	}

	log.Info().Msg(utils.GetDetectedMessage(diffResult.Files))
	for _, file := range diffResult.Files {
		log.Info().Msg("file: " + file)
	}

	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond, func(s *spinner.Spinner) {
		s.Prefix = "generate git message: "
	})
	fmt.Fprintln(os.Stderr, "→ generating commit message (timeout ~45s)...")
	s.Start()
	defer s.Stop()

	locale := "en"
	maxLength := 50
	if repoCfg.Commit.Locale != "" {
		locale = repoCfg.Commit.Locale
	}
	if repoCfg.Commit.MaxLength > 0 {
		maxLength = repoCfg.Commit.MaxLength
	}
	generatePrompt := utils.AppendAllowedTypes(
		utils.GeneratePrompt(locale, maxLength, utils.ConventionalCommitType),
		repoCfg.Commit.Types,
	)

	aiCtx, aiCancel := context.WithTimeout(ctx, 45*time.Second)
	defer aiCancel()

	aiDiff, compactStats := aiprovider.CompactDiffForAI(diffResult.Diff)
	if compactStats.Truncated {
		log.Warn().
			Int("original_bytes", compactStats.OriginalBytes).
			Int("compact_bytes", compactStats.CompactBytes).
			Int("files", compactStats.FileCount).
			Int("kept", compactStats.KeptFiles).
			Int("skipped", compactStats.SkippedFiles).
			Msg("diff too large for AI; sending abbreviated patch")
	}

	useCandidates := shouldUseCandidates(flags, repoCfg, params)
	msg, err := pickCommitMessage(ctx, aiCtx, params, flags, useCandidates, generatePrompt, aiDiff, diffResult.Diff, s)
	if err != nil {
		return err
	}
	if msg == "" {
		return nil
	}

	if err := enforceRepoPolicy(repoCfg, currentBranch(), msg, flags.skipPolicy); err != nil {
		return err
	}
	warnRepoPolicy(repoCfg, currentBranch(), msg)

	if err := squashQuickUpdates(ctx); err != nil {
		return err
	}
	if utils.IsDirty().Unwrap() {
		assert.Must(utils.ShellExec(ctx, "git", "add", "-A"))
	}

	fmt.Fprintln(os.Stderr, "→ committing...")
	if err := utils.GitCommit(ctx, msg); err != nil {
		return err
	}
	if err := ensurePushPolicy(repoRoot, utils.GetBranchName(), flags.overridePolicy); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "→ commit message: %s\n", msg)
	fmt.Fprintln(os.Stderr, "→ pushing to remote...")
	if err := finishPush(ctx); err != nil {
		return err
	}
	if flags.showPrompt && !useCandidates {
		fmt.Println("\n" + generatePrompt + "\n")
	}
	log.Info().Str("message", msg).Bool("candidates", useCandidates).Msg("commit message generated")
	workflow.PrintRecommendations(os.Stdout, "commit")
	return nil
}

func pickCommitMessage(
	ctx, aiCtx context.Context,
	params cmdParams,
	flags *flagOptions,
	useCandidates bool,
	generatePrompt, aiDiff, fullDiff string,
	s *spinner.Spinner,
) (string, error) {
	if useCandidates {
		candidates, err := aiprovider.GenerateCommitCandidates(aiCtx, params.AI, aiDiff)
		s.Stop()
		if err != nil {
			log.Warn().Err(err).Msg("AI candidates failed or timed out; using rule-based options")
		}
		if hint := aiprovider.BreakingChangeHint(fullDiff); hint != "" {
			log.Warn().Msg(hint)
			fmt.Println(hint)
		}
		options := make([]tap.SelectOption[string], 0, len(candidates))
		for _, candidate := range candidates {
			candidate := candidate
			options = append(options, tap.SelectOption[string]{
				Label: aiprovider.FormatCandidateLabel(candidate),
				Value: candidate.Message,
			})
		}
		if len(options) == 0 {
			return "", nil
		}
		if flags != nil && flags.candidates {
			fmt.Fprintln(os.Stderr, "→ pick a commit message (↑/↓ to move, Enter to confirm):")
			selected := tap.Select[string](ctx, tap.SelectOptions[string]{
				Message: "Pick a commit message:",
				Options: options,
			})
			return strings.TrimSpace(selected), nil
		}
		msg := aiprovider.AutoPickCandidate(candidates)
		fmt.Fprintf(os.Stderr, "→ commit message: %s\n", msg)
		if flags != nil && flags.edit {
			msg = strings.TrimSpace(tap.Text(ctx, tap.TextOptions{
				Message:      "git message(update or enter):",
				InitialValue: msg,
				DefaultValue: msg,
				Placeholder:  "update or enter",
			}))
		}
		return msg, nil
	}

	aiResp, err := params.AI.Complete(aiCtx, aiprovider.CompleteRequest{
		System: generatePrompt,
		User:   aiDiff,
	})
	s.Stop()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(aiCtx.Err(), context.DeadlineExceeded) {
			log.Warn().Msg("AI timed out; falling back to rule-based commit message")
			aiResp = aiprovider.CompleteResponse{
				Text:     aiprovider.CommitMessageFromDiff(fullDiff),
				Provider: "rule-fallback",
				Fallback: true,
			}
		} else {
			log.Err(err).Msg("failed to generate commit message")
			return "", errors.WrapCaller(err)
		}
	}

	if aiResp.Fallback {
		log.Warn().Str("provider", aiResp.Provider).Msg("using rule-based commit message fallback (AI unavailable)")
	}
	if hint := aiprovider.BreakingChangeHint(fullDiff); hint != "" {
		log.Warn().Msg(hint)
		fmt.Println(hint)
	}

	msg := strings.TrimSpace(aiResp.Text)
	fmt.Fprintf(os.Stderr, "→ commit message: %s\n", msg)
	if flags != nil && flags.edit {
		msg = strings.TrimSpace(tap.Text(ctx, tap.TextOptions{
			Message:      "git message(update or enter):",
			InitialValue: msg,
			DefaultValue: msg,
			Placeholder:  "update or enter",
		}))
	}
	return msg, nil
}

func finishPush(ctx context.Context) error {
	pushOut := utils.GitPush(ctx, "--force-with-lease", "origin", utils.GetBranchName())
	if shouldPullDueToRemoteUpdate(pushOut) {
		return handlePushRejected(ctx)
	}
	if strings.Contains(pushOut, "timed out") {
		return fmt.Errorf("push failed: %s", pushOut)
	}
	fmt.Fprintln(os.Stderr, "→ done")
	return nil
}

func squashQuickUpdates(ctx context.Context) error {
	prefixMsg := fmt.Sprintf("chore: quick update %s", utils.GetBranchName())
	targetCommit := getFirstNonPrefixCommit(ctx, prefixMsg)
	if targetCommit != "" {
		return utils.ShellExec(ctx, "git", "reset", "--soft", targetCommit)
	}

	commitsToSquash := getCommitsToSquash(ctx, prefixMsg)
	if len(commitsToSquash) == 0 {
		return nil
	}
	parentCommit := getParentCommit(ctx, commitsToSquash[0])
	if parentCommit != "" {
		return utils.ShellExec(ctx, "git", "reset", "--soft", parentCommit)
	}
	return utils.ShellExec(ctx, "git", "reset", "--soft", "HEAD~"+fmt.Sprint(len(commitsToSquash)))
}

func handlePushRejected(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "→ remote changed, pulling...")
	err := gitPull()
	if err != nil {
		if gitconflict.HasConflicts(ctx, "") {
			handleMergeConflict(ctx)
			return fmt.Errorf("push rejected; resolve conflicts then retry commit")
		}
		return fmt.Errorf("push rejected and pull failed: %w", err)
	}
	if gitconflict.HasConflicts(ctx, "") {
		handleMergeConflict(ctx)
		return fmt.Errorf("push rejected; resolve conflicts then retry commit")
	}
	fmt.Fprintln(os.Stderr, "→ retrying push...")
	pushOut := utils.GitPush(ctx, "--force-with-lease", "origin", utils.GetBranchName())
	if shouldPullDueToRemoteUpdate(pushOut) {
		return fmt.Errorf("push still rejected after pull; resolve manually and push again")
	}
	if strings.Contains(pushOut, "timed out") {
		return fmt.Errorf("push failed after pull: %s", pushOut)
	}
	fmt.Fprintln(os.Stderr, "→ done")
	return nil
}

func mustRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func shouldUseCandidates(flags *flagOptions, repoCfg repoconfig.Bundle, params cmdParams) bool {
	if flags != nil && flags.candidates {
		return true
	}
	if repoCfg.Commit.CandidatesDefault {
		return true
	}
	for _, cfg := range params.CommitCfg {
		if cfg != nil && cfg.CandidatesDefault {
			return true
		}
	}
	return false
}
