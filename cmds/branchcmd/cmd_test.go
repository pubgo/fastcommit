package branchcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckoutRemoteArgs(t *testing.T) {
	t.Run("bare branch name gets origin prefix", func(t *testing.T) {
		args, err := checkoutRemoteArgs("main")
		require.NoError(t, err)
		require.Equal(t, []string{"checkout", "-b", "main", "--track", "origin/main"}, args)
	})

	t.Run("nested branch name keeps local path", func(t *testing.T) {
		args, err := checkoutRemoteArgs("origin/feature/demo")
		require.NoError(t, err)
		require.Equal(t, []string{"checkout", "-b", "feature/demo", "--track", "origin/feature/demo"}, args)
	})

	t.Run("empty name is rejected", func(t *testing.T) {
		_, err := checkoutRemoteArgs("")
		require.Error(t, err)
	})
}

func TestRequireOneArg(t *testing.T) {
	t.Run("single arg passes", func(t *testing.T) {
		name, err := requireOneArg([]string{"feature"}, "branch create <name>")
		require.NoError(t, err)
		require.Equal(t, "feature", name)
	})

	t.Run("no arg fails with usage", func(t *testing.T) {
		_, err := requireOneArg(nil, "branch create <name>")
		require.ErrorContains(t, err, "usage: branch create <name>")
	})

	t.Run("multiple args fail", func(t *testing.T) {
		_, err := requireOneArg([]string{"a", "b"}, "branch create <name>")
		require.Error(t, err)
	})
}
