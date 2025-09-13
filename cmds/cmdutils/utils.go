package cmdutils

import (
	"sync"

	"github.com/pubgo/fastcommit/configs"
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/log"
)

var GetBranchName = sync.OnceValue(func() string { return utils.GetCurrentBranch().Must() })

func LoadConfigAndBranch() {
	branchName := GetBranchName()
	log.Info().Msg("current branch: " + branchName)
	log.Info().Msg("config: " + configs.GetConfigPath())
}
