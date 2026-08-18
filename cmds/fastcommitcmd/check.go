package fastcommitcmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pubgo/fastgit/cmds/checkcmd"
	"github.com/pubgo/fastgit/pkg/repoconfig"
)

const preCommitCheckTimeout = 10 * time.Minute

func runPreCommitCheck(ctx context.Context, repoRoot string, skip bool, autoFix bool) error {
	if skip {
		return nil
	}

	cfg := checkcmd.ForCommit(checkcmd.LoadConfig(repoRoot))
	if cfg == nil {
		return nil
	}

	fmt.Fprintf(os.Stderr, "→ running pre-commit check (%s)...\n", checkcmd.ConfigPath(repoRoot))
	checkCtx, cancel := context.WithTimeout(ctx, preCommitCheckTimeout)
	defer cancel()

	_, err := checkcmd.Run(checkCtx, *cfg, checkcmd.RunOptions{
		StagedOnly: true,
		Fix:        autoFix,
		RepoRoot:   repoRoot,
	})
	if err != nil {
		if checkCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("pre-commit check timed out after %s\nhint: fix slow tests, or use --skip-check to bypass", preCommitCheckTimeout)
		}
		return fmt.Errorf("pre-commit check failed: %w\nhint: fix issues, or use --skip-check to bypass", err)
	}
	fmt.Fprintln(os.Stderr, "→ pre-commit check passed")
	return nil
}

func ensurePushPolicy(repoRoot, branch string, override bool) error {
	cfg, err := repoconfig.Load(repoRoot)
	if err != nil {
		return err
	}
	return cfg.ValidatePush(branch, override)
}
