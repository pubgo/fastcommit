package tagcmd

import (
	"os"
	"path/filepath"
	"testing"

	semver "github.com/hashicorp/go-version"
	"github.com/pubgo/fastgit/cmds/fastcommitcmd"
	"github.com/stretchr/testify/require"
)

func TestEnsureVersionAligned(t *testing.T) {
	tmp := t.TempDir()
	verFile := filepath.Join(tmp, "VERSION")
	require.NoError(t, os.WriteFile(verFile, []byte("v1.2.3\n"), 0o644))

	tag := semver.Must(semver.NewVersion("v1.2.3"))
	err := ensureVersionAligned(verFile, tag, []*fastcommitcmd.Config{{GenVersion: true}})
	require.NoError(t, err)
}

func TestEnsureVersionAlignedMismatch(t *testing.T) {
	tmp := t.TempDir()
	verFile := filepath.Join(tmp, "VERSION")
	require.NoError(t, os.WriteFile(verFile, []byte("v1.2.3\n"), 0o644))

	tag := semver.Must(semver.NewVersion("v1.2.4"))
	err := ensureVersionAligned(verFile, tag, []*fastcommitcmd.Config{{GenVersion: true}})
	require.Error(t, err)
}

func TestShowTagArgs(t *testing.T) {
	t.Run("tag name builds git show args", func(t *testing.T) {
		args, err := showTagArgs("v0.0.15")
		require.NoError(t, err)
		require.Equal(t, []string{"show", "v0.0.15"}, args)
	})

	t.Run("empty tag name is rejected", func(t *testing.T) {
		_, err := showTagArgs("   ")
		require.Error(t, err)
	})
}
