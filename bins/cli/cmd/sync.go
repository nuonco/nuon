package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/services/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/services/version"
)

func (c *cli) syncCmd() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync local config files to Nuon",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			// Reuse the existing sync functionality from apps
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.SyncDir(cmd.Context(), ".", version.Version)
		}),
		GroupID: CoreGroup.ID,
	}

	return syncCmd
}
