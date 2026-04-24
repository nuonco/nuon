package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

func (c *cli) syncCmd() *cobra.Command {
	var create bool
	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync local config files to Nuon",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			if create {
				return svc.SyncDirWithCreate(cmd.Context(), ".", version.Version)
			}
			return svc.SyncDir(cmd.Context(), ".", version.Version)
		}),
		GroupID: CoreGroup.ID,
	}
	syncCmd.Flags().BoolVar(&create, "create", false, "Create the app if it doesn't exist")

	return syncCmd
}
