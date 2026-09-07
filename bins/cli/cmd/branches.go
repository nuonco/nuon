package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
)

func (c *cli) branchesCmd() *cobra.Command {
	cmd := c.newBranchesCmd(false)
	cmd.GroupID = CoreGroup.ID
	return cmd
}

func (c *cli) newBranchesCmd(deprecatedAlias bool) *cobra.Command {
	var (
		appID    string
		branchID string
	)

	branchesCmd := &cobra.Command{
		Use:               "branches",
		Short:             "Manage app branches",
		Aliases:           []string{"br"},
		PersistentPreRunE: c.persistentPreRunE,
	}
	if deprecatedAlias {
		branchesCmd.Deprecated = "use `nuon branches` instead"
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List branches for an app",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.ListBranches(cmd.Context(), appID, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	branchesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get branch details",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.GetBranch(cmd.Context(), appID, branchID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	getCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	getCmd.MarkFlagRequired("branch-id")
	branchesCmd.AddCommand(getCmd)

	var branchName string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new branch",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.CreateBranch(cmd.Context(), appID, branchName, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	createCmd.Flags().StringVarP(&branchName, "name", "n", "", "Branch name")
	createCmd.MarkFlagRequired("name")
	branchesCmd.AddCommand(createCmd)

	var (
		force  bool
		noWait bool
	)
	triggerCmd := &cobra.Command{
		Use:         "trigger",
		Short:       "Trigger a branch run",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			opts := apps.TriggerBranchRunOptions{
				Force:  force,
				NoWait: noWait,
			}
			return c.apps.TriggerBranchRun(cmd.Context(), appID, branchID, opts, PrintJSON)
		}),
	}
	triggerCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	triggerCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	triggerCmd.Flags().BoolVar(&force, "force", false, "Force rebuild all components")
	triggerCmd.Flags().BoolVar(&noWait, "no-wait", false, "Return immediately after triggering without launching the workflow viewer")
	branchesCmd.AddCommand(triggerCmd)

	var (
		previewConfigID    string
		previewPRNumber    int
		previewGitRef      string
		previewHeadSHA     string
		previewInstallID   string
		previewMode        string
		previewForce       bool
		previewAutoApprove bool
		previewWait        bool
		previewNoWait      bool
	)
	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Trigger a git preview run against a PR or branch",
		Long: `Trigger a git preview run against a pull request or branch.

By default, an interactive wizard selects the app branch, preview mode, source,
and installation. Use flags with --output json or --output agent for scripting.`,
		Annotations: annotations(tuiAnnotation(TUIAltScreen), outputsAnnotation(OutputTable, OutputJSON, OutputAgent)),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			opts := apps.PreviewBranchRunOptions{
				ConfigID:    previewConfigID,
				GitRef:      previewGitRef,
				HeadSHA:     previewHeadSHA,
				InstallID:   previewInstallID,
				Mode:        previewMode,
				Force:       previewForce,
				AutoApprove: previewAutoApprove,
				Wait:        previewWait,
				NoWait:      previewNoWait,
			}
			if cmd.Flags().Changed("pr-number") {
				opts.PRNumber = &previewPRNumber
			}
			return c.apps.PreviewBranchRun(cmd.Context(), appID, branchID, opts, PrintJSON)
		}),
	}
	previewCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	previewCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	previewCmd.Flags().IntVar(&previewPRNumber, "pr-number", 0, "Pull request number to preview")
	previewCmd.Flags().StringVar(&previewGitRef, "git-ref", "", "Git branch to preview")
	previewCmd.Flags().StringVar(&previewHeadSHA, "head-sha", "", "Commit SHA for the preview source")
	previewCmd.Flags().StringVar(&previewInstallID, "install-id", "", "Install to run the preview against")
	previewCmd.Flags().StringVar(&previewMode, "mode", "", "Preview mode: plan-only, apply, or build-only")
	previewCmd.Flags().StringVar(&previewConfigID, "config-id", "", "Branch config ID (defaults to latest)")
	previewCmd.Flags().BoolVar(&previewForce, "force", false, "Force rebuild all components")
	previewCmd.Flags().BoolVar(&previewAutoApprove, "auto-approve", false, "Skip the approval gate before deploy steps")
	previewCmd.Flags().BoolVar(&previewWait, "wait", false, "Block until the preview workflow completes")
	previewCmd.Flags().BoolVar(&previewNoWait, "no-wait", false, "Return after triggering without opening the workflow viewer")
	branchesCmd.AddCommand(previewCmd)

	var confirmDelete bool
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an app branch",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.DeleteBranch(cmd.Context(), appID, branchID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	deleteCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm deletion")
	deleteCmd.MarkFlagRequired("branch-id")
	deleteCmd.MarkFlagRequired("confirm")
	branchesCmd.AddCommand(deleteCmd)

	runsCmd := &cobra.Command{
		Use:         "runs",
		Short:       "List branch runs and monitor a selected workflow",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.ListBranchRuns(cmd.Context(), appID, branchID, PrintJSON)
		}),
	}
	runsCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	runsCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	branchesCmd.AddCommand(runsCmd)

	var (
		syncPath    string
		syncConfirm bool
		syncDryRun  bool
	)
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Reconcile app branches from standalone TOML files",
		Long: `Reconcile app branches from a TOML file or directory of TOML files.

Each file is one standalone branch config (the same shape as branch.toml).
Directory mode recursively loads sorted *.toml files, rejects duplicate names,
and proposes deleting config-managed remote branches that are absent locally.
A single file reconciles only that named branch and never deletes others.

--app-id defaults to the app selected with nuon apps select.`,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.SyncBranches(cmd.Context(), apps.SyncBranchesOptions{
				Path:      syncPath,
				AppID:     appID,
				Confirm:   syncConfirm,
				DryRun:    syncDryRun,
				PrintJSON: PrintJSON,
			})
		}),
	}
	syncCmd.Flags().StringVarP(&syncPath, "file", "d", "", "Branch TOML file or directory of TOML files")
	syncCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	syncCmd.Flags().BoolVar(&syncConfirm, "confirm", false, "Apply the plan without prompting")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Print the plan without applying changes")
	syncCmd.MarkFlagRequired("file")
	branchesCmd.AddCommand(syncCmd)

	return branchesCmd
}
