package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestBundleCommandSurface(t *testing.T) {
	c := &cli{}
	root := &cobra.Command{Use: "nuon"}
	apps := &cobra.Command{Use: "apps"}
	root.AddCommand(apps)
	group := c.appBundlesCmd()
	apps.AddCommand(group)
	require.Len(t, group.Commands(), 5)
	previous := ReadOnly
	t.Cleanup(func() { ReadOnly = previous })
	ReadOnly = true
	for _, name := range []string{"list", "get", "create", "wait", "download"} {
		cmd, _, err := group.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, cmd.Name())
		require.Nil(t, cmd.Flags().Lookup("no-wait"))
		require.NotNil(t, cmd.Run)
		require.NotEmpty(t, cmd.Annotations)
		if name == "create" {
			require.Error(t, guardReadOnly(cmd))
		} else {
			require.NoError(t, guardReadOnly(cmd))
		}
		if name == "download" {
			require.NotNil(t, cmd.Flags().Lookup("file"))
			require.Nil(t, cmd.Flags().Lookup("output"))
		}
	}
	found, remaining, err := group.Find([]string{"export"})
	require.NoError(t, err)
	require.Same(t, group, found)
	require.Equal(t, []string{"export"}, remaining)
	other := &cobra.Command{Use: "download"}
	root.AddCommand(other)
	require.Error(t, guardReadOnly(other))
}
