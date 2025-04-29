package cmdutils

import (
	"strings"

	"github.com/pubgo/funk/log"

	"github.com/pubgo/fastcommit/configs"
)

func LoadConfigAndBranch() {
	branchName := configs.GetBranchName()
	log.Info().Msg("current branch: " + strings.TrimSpace(branchName))
	log.Info().Msg("config: " + configs.GetConfigPath())
}
