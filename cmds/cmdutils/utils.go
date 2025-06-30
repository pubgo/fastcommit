package cmdutils

import (
	"github.com/pubgo/fastcommit/utils"
	"github.com/pubgo/funk/assert"
	"strings"

	"github.com/pubgo/funk/log"

	"github.com/pubgo/fastcommit/configs"
)

var curBranchName string

func GetBranchName() string {
	if curBranchName != "" {
		return curBranchName
	}

	curBranchName = assert.Exit1(utils.RunOutput("git", "rev-parse", "--abbrev-ref", "HEAD"))
	return curBranchName
}

func LoadConfigAndBranch() {
	branchName := GetBranchName()
	log.Info().Msg("current branch: " + strings.TrimSpace(branchName))
	log.Info().Msg("config: " + configs.GetConfigPath())
}
