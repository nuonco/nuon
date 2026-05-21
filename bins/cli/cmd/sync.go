package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

func (c *cli) syncCmd() *cobra.Command {
	var (
		create bool
		appID  string
	)
	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync local config files to Nuon",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			opts := apps.SyncOptions{
				AppFlag: appID,
				Create:  create,
			}
			svc := apps.New(c.v, c.apiClient, c.cfg)
			if create {
				return svc.SyncDirWithCreate(cmd.Context(), ".", version.Version, opts)
			}
			return svc.SyncDir(cmd.Context(), ".", version.Version, opts)
		}),
		GroupID: CoreGroup.ID,
	}
	syncCmd.Flags().BoolVar(&create, "create", false, "Create the app if it doesn't exist")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app (default: selected app or cwd dir name)")

	return syncCmd
}
