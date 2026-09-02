package remotecmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteArgs(t *testing.T) {
	require.Equal(t, []string{"remote", "-v"}, remoteArgs())
}
