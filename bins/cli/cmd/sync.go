package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

// syncLongHelp documents the sync → build-wait phases and exit codes; shared
// by `nuon sync` and `nuon apps sync`.
const syncLongHelp = `Sync local config files to Nuon.

When the synced config changes components, the sync schedules component builds
and waits for them to complete (up to 20m), reporting each build's outcome.
Use --no-wait to skip waiting; the exit code then reflects the sync only.

Exit codes:
  0 - config synced (and all scheduled builds completed, unless --no-wait)
  1 - sync failed
  3 - config synced, but one or more component builds failed, were
      policy-blocked, or timed out`

func (c *cli) syncCmd() *cobra.Command {
	var (
		create    bool
		force     bool
		appID     string
		branch    string
		appBranch bool
		preview   bool
		noWait    bool
	)
	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync local config files to Nuon",
		Long:              syncLongHelp,
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			opts := apps.SyncOptions{
				AppFlag:   appID,
				Force:     force,
				Create:    create,
				Branch:    branch,
				AppBranch: appBranch,
				Preview:   preview,
				PrintJSON: PrintJSON,
				NoWait:    noWait,
			}
			svc := c.apps
			if create {
				return svc.SyncDirWithCreate(cmd.Context(), ".", version.Version, opts)
			}
			return svc.SyncDir(cmd.Context(), ".", version.Version, opts)
		}),
		GroupID: CoreGroup.ID,
	}
	syncCmd.Flags().BoolVar(&create, "create", false, "Create the app if it doesn't exist")
	syncCmd.Flags().BoolVar(&force, "force", false, "Sync to the configured app even if the directory name does not match")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to sync this config with (defaults to the selected app)")
	syncCmd.Flags().StringVar(&branch, "branch", "", "Target a specific app branch for this sync")
	syncCmd.Flags().BoolVar(&appBranch, "app-branch", false, "Select an app branch interactively and trigger a branch run after sync")
	syncCmd.Flags().BoolVar(&preview, "preview", false, "Plan-only preview mode (no apply). Only used with --branch or --app-branch")
	syncCmd.Flags().BoolVar(&noWait, "no-wait", false, "Do not wait for scheduled component builds to complete")

	return syncCmd
}
