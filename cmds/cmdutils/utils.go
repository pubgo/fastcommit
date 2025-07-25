package cmdutils

import (
	"strings"

	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/log"
)

var GetBranchName = utils.Once(func() string { return utils.GetCurrentBranch().Must() })

func LoadConfigAndBranch() {
	branchName := GetBranchName()
	log.Info().Msg("current branch: " + strings.TrimSpace(branchName))
	log.Info().Msg("config: " + configs.GetConfigPath())
}
