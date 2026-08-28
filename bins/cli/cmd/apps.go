package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

func (c *cli) appsCmd() *cobra.Command {
	var (
		noSelect bool
		offset   int
		limit    int
	)

	appsCmd := &cobra.Command{
		Use:               "apps",
		Short:             "Manage apps",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           CoreGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.List(cmd.Context(), offset, limit, PrintJSON)
		}),
	}

	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	appsCmd.AddCommand(listCmd)

	var (
		bundleAppID     string
		bundleOffset    int
		bundleLimit     int
		bundleFile      string
		bundleNoResume  bool
		bundleOverwrite bool
		bundleConfigID  string
		bundlePlatform  string
		bundleNoWait    bool
	)
	bundlesCmd := &cobra.Command{
		Use:   "bundles",
		Short: "View and download published portable bundles",
	}
	bundlesCreateCmd := &cobra.Command{
		Use:         "create",
		Short:       "Create and publish a portable bundle",
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.CreateBundle(cmd.Context(), bundleAppID, bundleConfigID, bundlePlatform, apps.CreateBundleOptions{
				NoWait:    bundleNoWait,
				PrintJSON: PrintJSON,
			})
		}),
	}
	bundlesCreateCmd.Flags().StringVarP(&bundleAppID, "app-id", "a", "", "The ID or name of an app (defaults to the selected app)")
	bundlesCreateCmd.Flags().StringVar(&bundleConfigID, "config-id", "", "Exact app config ID")
	bundlesCreateCmd.Flags().StringVar(&bundlePlatform, "platform", "linux/amd64", "Target platform")
	bundlesCreateCmd.Flags().BoolVar(&bundleNoWait, "no-wait", false, "Return immediately without waiting for the bundle to become active")
	bundlesCreateCmd.MarkFlagRequired("config-id")
	bundlesCmd.AddCommand(bundlesCreateCmd)
	bundlesListCmd := &cobra.Command{
		Use:         "list",
		Aliases:     []string{"ls"},
		Short:       "List published portable bundles",
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.ListBundles(cmd.Context(), bundleAppID, bundleOffset, bundleLimit, PrintJSON)
		}),
	}
	bundlesListCmd.Flags().StringVarP(&bundleAppID, "app-id", "a", "", "The ID or name of an app (defaults to the selected app)")
	bundlesListCmd.Flags().IntVar(&bundleOffset, "offset", 0, "Offset for pagination")
	bundlesListCmd.Flags().IntVar(&bundleLimit, "limit", 20, "Limit for pagination")
	bundlesCmd.AddCommand(bundlesListCmd)

	bundlesGetCmd := &cobra.Command{
		Use:         "get <bundle-id>",
		Short:       "Get a published portable bundle",
		Args:        cobra.ExactArgs(1),
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.GetBundle(cmd.Context(), bundleAppID, args[0], PrintJSON)
		}),
	}
	bundlesGetCmd.Flags().StringVarP(&bundleAppID, "app-id", "a", "", "The ID or name of an app (defaults to the selected app)")
	bundlesCmd.AddCommand(bundlesGetCmd)

	bundlesDownloadCmd := &cobra.Command{
		Use:         "download <bundle-id>",
		Short:       "Download a published portable bundle",
		Long:        "Download a published portable bundle to the path specified by --file. The global --output flag remains reserved for output formatting.",
		Args:        cobra.ExactArgs(1),
		Annotations: outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.DownloadBundle(cmd.Context(), bundleAppID, args[0], apps.DownloadBundleOptions{
				File:      bundleFile,
				NoResume:  bundleNoResume,
				Overwrite: bundleOverwrite,
			})
		}),
	}
	bundlesDownloadCmd.Flags().StringVarP(&bundleAppID, "app-id", "a", "", "The ID or name of an app (defaults to the selected app)")
	bundlesDownloadCmd.Flags().StringVar(&bundleFile, "file", "", "Destination file path (required; --output controls CLI formatting)")
	bundlesDownloadCmd.Flags().BoolVar(&bundleNoResume, "no-resume", false, "Restart instead of resuming an existing partial download")
	bundlesDownloadCmd.Flags().BoolVar(&bundleOverwrite, "overwrite", false, "Replace an existing destination file")
	bundlesDownloadCmd.MarkFlagRequired("file")
	bundlesCmd.AddCommand(bundlesDownloadCmd)

	var bundleCacheDir string
	bundlesPullCmd := &cobra.Command{
		Use:         "pull <bundle-id>",
		Short:       "Differentially download a published portable bundle",
		Long:        "Download only the bundle blobs missing from the local cache, then reassemble the exact bundle archive at --file. Repeat pulls of similar bundles transfer only what changed.",
		Args:        cobra.ExactArgs(1),
		Annotations: outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.PullBundle(cmd.Context(), bundleAppID, args[0], apps.PullBundleOptions{
				File:      bundleFile,
				CacheDir:  bundleCacheDir,
				Overwrite: bundleOverwrite,
			})
		}),
	}
	bundlesPullCmd.Flags().StringVarP(&bundleAppID, "app-id", "a", "", "The ID or name of an app (defaults to the selected app)")
	bundlesPullCmd.Flags().StringVar(&bundleFile, "file", "", "Destination file path (required; --output controls CLI formatting)")
	bundlesPullCmd.Flags().StringVar(&bundleCacheDir, "cache-dir", "", "Blob cache directory (defaults to ~/.config/nuon/bundle-blobs)")
	bundlesPullCmd.Flags().BoolVar(&bundleOverwrite, "overwrite", false, "Replace an existing destination file")
	bundlesPullCmd.MarkFlagRequired("file")
	bundlesCmd.AddCommand(bundlesPullCmd)
	appsCmd.AddCommand(bundlesCmd)

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an app",
		Long:  "Get either the current app or an app by name or ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Get(cmd.Context(), appID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	currentCmd := &cobra.Command{
		Use:    "current",
		Short:  "Get the current app (deprecated)",
		Hidden: true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			printDeprecatedCommandWarning(cmd, "Use `nuon apps get` instead")

			svc := c.apps
			return svc.Get(cmd.Context(), c.cfg.GetString("app_id"), PrintJSON)
		}),
	}
	appsCmd.AddCommand(currentCmd)

	latestSandboxConfigCmd := &cobra.Command{
		Use:   "sandbox-config",
		Short: "View sandbox config",
		Long:  "View apps latest sandbox config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.GetSandboxConfig(cmd.Context(), appID, PrintJSON)
		}),
	}
	latestSandboxConfigCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestSandboxConfigCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestSandboxConfigCmd)

	configs := &cobra.Command{
		Use:   "configs",
		Short: "List app configs",
		Long:  "List app configs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.ListConfigs(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	configs.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	configs.MarkFlagRequired("app-id")
	configs.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	configs.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	appsCmd.AddCommand(configs)

	latestInputConfig := &cobra.Command{
		Use:   "input-config",
		Short: "View app input config",
		Long:  "View latest app input config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.GetInputConfig(cmd.Context(), appID, PrintJSON)
		}),
	}
	latestInputConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestInputConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestInputConfig)

	latestRunnerConfig := &cobra.Command{
		Use:   "runner-config",
		Short: "View app runner config",
		Long:  "View latest app runner config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.GetRunnerConfig(cmd.Context(), appID, PrintJSON)
		}),
	}
	latestRunnerConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestRunnerConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestRunnerConfig)

	selectAppCmd := &cobra.Command{
		Use:         "select",
		Short:       "Select your current app",
		Long:        "Select your current app from a list or by app ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Select(cmd.Context(), appID, PrintJSON)
		}),
	}
	selectAppCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	appsCmd.AddCommand(selectAppCmd)

	deselectAppCmd := &cobra.Command{
		Use:   "deselect",
		Short: "Deselect your current app",
		Long:  "Deselect your current app",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Deselect(cmd.Context())
		}),
	}
	appsCmd.AddCommand(deselectAppCmd)

	unsetCurrentAppCmd := &cobra.Command{
		Use:        "unset-current",
		Deprecated: "Use `nuon apps deselect` instead",
		Short:      "Unset your current app (deprecated)",
		Hidden:     true,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Deselect(cmd.Context())
		}),
	}
	appsCmd.AddCommand(unsetCurrentAppCmd)

	var (
		syncCreate      bool
		syncForce       bool
		syncAppID       string
		syncBranch      string
		syncAppBranch   bool
		syncPreview     bool
		syncAutoApprove bool
		syncNoWait      bool
	)
	syncCmd := &cobra.Command{
		Use:               "sync [dir]",
		Short:             "Sync nuon app directory",
		Long:              syncLongHelp,
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			var dirName string
			if len(args) > 0 {
				dirName = args[0]
			} else {
				var err error
				dirName, err = os.Getwd()
				if err != nil {
					return errors.Wrap(err, "unable to get directory name")
				}
			}

			opts := apps.SyncOptions{
				AppFlag:     syncAppID,
				Force:       syncForce,
				Create:      syncCreate,
				Branch:      syncBranch,
				AppBranch:   syncAppBranch,
				Preview:     syncPreview,
				AutoApprove: syncAutoApprove,
				PrintJSON:   PrintJSON,
				NoWait:      syncNoWait,
			}
			svc := c.apps
			if syncCreate {
				return svc.SyncDirWithCreate(cmd.Context(), dirName, version.Version, opts)
			}
			return svc.SyncDir(cmd.Context(), dirName, version.Version, opts)
		}),
	}
	syncCmd.Flags().BoolVar(&syncCreate, "create", false, "Create the app if it doesn't exist")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Sync to the configured app even if the directory name does not match")
	syncCmd.Flags().StringVarP(&syncAppID, "app-id", "a", "", "The ID or name of the app to sync this config with (defaults to the selected app)")
	syncCmd.Flags().StringVar(&syncBranch, "branch", "", "Target a specific app branch for this sync")
	syncCmd.Flags().BoolVar(&syncAppBranch, "app-branch", false, "Select an app branch interactively and trigger a branch run after sync")
	syncCmd.Flags().BoolVar(&syncPreview, "preview", false, "Plan-only preview mode (no apply). Only used with --branch or --app-branch")
	syncCmd.Flags().BoolVar(&syncAutoApprove, "auto-approve", false, "Skip the branch run's approval gate before each install group deploys")
	syncCmd.Flags().BoolVar(&syncNoWait, "no-wait", false, "Do not wait for scheduled component builds to complete")
	appsCmd.AddCommand(syncCmd)

	var (
		buildAppID    string
		buildConfigID string
	)
	buildCmd := &cobra.Command{
		Use:         "build",
		Short:       "Build all components for an app config [preview]",
		Long:        "Triggers a workflow that builds all components defined in the app config. If no config ID is provided, uses the latest config.",
		Hidden:      true,
		Annotations: tuiAnnotation(TUIAltScreen),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := c.persistentPreRunE(cmd, args); err != nil {
				return err
			}
			if !c.cfg.Preview {
				return fmt.Errorf("[NUON_PREVIEW=false] apps build is a preview feature, set NUON_PREVIEW=true to enable")
			}
			return nil
		},
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Build(cmd.Context(), buildAppID, buildConfigID)
		}),
	}
	buildCmd.Flags().StringVarP(&buildAppID, "app-id", "a", "", "The ID or name of an app (default: current app)")
	buildCmd.Flags().StringVar(&buildConfigID, "config-id", "", "The config ID to build (default: latest)")
	appsCmd.AddCommand(buildCmd)

	syncDirCmd := &cobra.Command{
		Deprecated:        "use `nuon sync` instead",
		Use:               "sync-dir",
		Short:             "Sync nuon app directory (deprecated)",
		Hidden:            true,
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			var dirName string
			if len(args) > 0 {
				dirName = args[0]
			} else {
				var err error
				dirName, err = os.Getwd()
				if err != nil {
					return errors.Wrap(err, "unable to get directory name")
				}
			}

			svc := c.apps
			return svc.DeprecatedSyncDir(cmd.Context(), dirName, version.Version, apps.SyncOptions{})
		}),
	}
	appsCmd.AddCommand(syncDirCmd)

	validateCmd := &cobra.Command{
		Use:               "validate",
		Short:             "Validate nuon app directory",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			var dirName string
			if len(args) > 0 {
				dirName = args[0]
			} else {
				var err error
				dirName, err = os.Getwd()
				if err != nil {
					return errors.Wrap(err, "unable to get directory name")
				}
			}

			svc := c.apps
			return svc.ValidateDir(cmd.Context(), dirName)
		}),
	}
	appsCmd.AddCommand(validateCmd)

	var name string
	createCmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a new app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Create(cmd.Context(), name, PrintJSON, noSelect)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "app name")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "do not automatically set the new app as the current app")

	appsCmd.AddCommand(createCmd)

	// nuon apps delete
	var confirmDelete bool
	deleteCmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete an existing app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Delete(cmd.Context(), appID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the app")
	deleteCmd.MarkFlagRequired("app-id")
	deleteCmd.MarkFlagRequired("confirm")

	appsCmd.AddCommand(deleteCmd)

	// nuon app generate/init commandasss
	appsCmd.AddCommand(c.initCmd())

	var rename bool
	renameCmd := &cobra.Command{
		Use:               "rename",
		Short:             "Rename an app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Rename(cmd.Context(), appID, name, rename, PrintJSON)
		}),
	}
	renameCmd.Flags().StringVarP(&name, "name", "n", "", "app name")
	renameCmd.MarkFlagRequired("name")
	renameCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	renameCmd.MarkFlagRequired("app-id")
	renameCmd.Flags().BoolVarP(&rename, "rename", "", true, "Rename config file if it exists")

	appsCmd.AddCommand(renameCmd)

	// variables subcommand (replacing secrets)
	variablesCmd := c.variablesCmd()
	appsCmd.AddCommand(variablesCmd)

	// branches subcommand
	branchesCmd := c.branchesCmd()
	appsCmd.AddCommand(branchesCmd)

	return appsCmd
}

func (c *cli) branchesCmd() *cobra.Command {
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

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List branches for an app",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.ListBranches(cmd.Context(), appID, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app. Defaults to the selected app.")
	branchesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get branch details",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.GetBranch(cmd.Context(), appID, branchID, PrintJSON)
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
			svc := c.apps
			return svc.CreateBranch(cmd.Context(), appID, branchName, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	createCmd.Flags().StringVarP(&branchName, "name", "n", "", "Branch name")
	createCmd.MarkFlagRequired("app-id")
	createCmd.MarkFlagRequired("name")
	branchesCmd.AddCommand(createCmd)

	var (
		planOnly   bool
		force      bool
		noWait     bool
		prNumber   int
		headSHA    string
		baseBranch string
	)
	triggerCmd := &cobra.Command{
		Use:         "trigger",
		Short:       "Trigger a branch run",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			opts := apps.TriggerBranchRunOptions{
				PlanOnly:   planOnly,
				Force:      force,
				NoWait:     noWait,
				HeadSHA:    headSHA,
				BaseBranch: baseBranch,
			}
			if cmd.Flags().Changed("pr-number") {
				opts.PRNumber = &prNumber
			}
			return svc.TriggerBranchRun(cmd.Context(), appID, branchID, opts, PrintJSON)
		}),
	}
	triggerCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	triggerCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	triggerCmd.Flags().BoolVar(&planOnly, "preview", false, "Plan-only preview mode (no apply)")
	triggerCmd.Flags().BoolVar(&force, "force", false, "Force rebuild all components")
	triggerCmd.Flags().BoolVar(&noWait, "no-wait", false, "Return immediately after triggering without launching the workflow viewer")
	triggerCmd.Flags().IntVar(&prNumber, "pr-number", 0, "Pull request number this run belongs to. Required for the run to comment on the PR.")
	triggerCmd.Flags().StringVar(&headSHA, "head-sha", "", "Commit SHA at the head of the pull request, used to set the commit status")
	triggerCmd.Flags().StringVar(&baseBranch, "base-branch", "", "Base branch the pull request targets")
	branchesCmd.AddCommand(triggerCmd)

	var confirmDelete bool
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an app branch",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.DeleteBranch(cmd.Context(), appID, branchID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	deleteCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm deletion")
	deleteCmd.MarkFlagRequired("app-id")
	deleteCmd.MarkFlagRequired("branch-id")
	deleteCmd.MarkFlagRequired("confirm")
	branchesCmd.AddCommand(deleteCmd)

	runsCmd := &cobra.Command{
		Use:         "runs",
		Short:       "List branch runs and monitor a selected workflow",
		Annotations: tuiAnnotation(TUIAltScreen),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.ListBranchRuns(cmd.Context(), appID, branchID, PrintJSON)
		}),
	}
	runsCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	runsCmd.Flags().StringVarP(&branchID, "branch-id", "b", "", "The ID or name of the branch")
	branchesCmd.AddCommand(runsCmd)

	return branchesCmd
}

func (c *cli) variablesCmd() *cobra.Command {
	var (
		appID      string
		variableID string
		offset     int
		limit      int
	)

	variablesCmd := &cobra.Command{
		Use:               "variables",
		Short:             "Create and manage app variables.",
		PersistentPreRunE: c.persistentPreRunE,
	}

	// list command
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all app variables",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.variables
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app")
	listCmd.MarkFlagRequired("app-id")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "The offset to start listing variables from")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "The number of variables to list")
	variablesCmd.AddCommand(listCmd)

	// delete command
	confirmDelete := false
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an app variable",
		Long:  "Delete an app variable value",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.variables
			return svc.Delete(cmd.Context(), appID, variableID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app")
	deleteCmd.Flags().StringVarP(&variableID, "variable-id", "i", "", "The ID or name of the variable to delete")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the variable")

	deleteCmd.MarkFlagRequired("app-id")
	deleteCmd.MarkFlagRequired("variable-id")
	deleteCmd.MarkFlagRequired("confirm")
	variablesCmd.AddCommand(deleteCmd)

	// create command
	var (
		name  string
		value string
	)
	createCmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a new app variable.",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.variables
			return svc.Create(cmd.Context(), appID, name, value, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name of the variable, must be alphanumeric, lower case and can use underscores.")
	createCmd.Flags().StringVarP(&value, "value", "v", "", "The variable value.")
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("value")
	createCmd.MarkFlagRequired("app-id")
	variablesCmd.AddCommand(createCmd)

	return variablesCmd
}
