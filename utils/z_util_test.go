package utils_test

import (
	"testing"

	"github.com/pubgo/fastcommit/utils"
	"github.com/stretchr/testify/assert"
)

func TestErrTagExists(t *testing.T) {
	var errMsg = `
To github.com:pubgo/funk.git
 ! [rejected]          v0.5.69-alpha.23 -> v0.5.69-alpha.23 (already exists)
error: failed to push some refs to 'github.com:pubgo/funk.git'
hint: Updates were rejected because the tag already exists in the remote.`
	assert.Equal(t, utils.IsRemoteTagExist(errMsg), true)
}
