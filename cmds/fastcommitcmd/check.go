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

func runPreCommitCheck(ctx context.Context, repoRoot string, skip bool) error {
	if skip {
		return nil
	}

	fmt.Fprintln(os.Stderr, "→ running pre-commit check...")
	checkCtx, cancel := context.WithTimeout(ctx, preCommitCheckTimeout)
	defer cancel()

	cfg := checkcmd.ForCommit(checkcmd.LoadConfig(repoRoot))
	_, err := checkcmd.Run(checkCtx, cfg, checkcmd.RunOptions{
		StagedOnly: true,
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
