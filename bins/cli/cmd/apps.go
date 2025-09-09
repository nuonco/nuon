package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/variables"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
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
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.List(cmd.Context(), offset, limit, PrintJSON)
		}),
	}

	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	appsCmd.AddCommand(listCmd)

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an app",
		Long:  "Get either the current app or an app by name or ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), appID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	currentCmd := &cobra.Command{
		Deprecated: "Use `nuon apps get` instead",
		Short:      "Get the current app (deprecated)",
		Use:        "current",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), c.cfg.GetString("app_id"), PrintJSON)
		}),
	}
	appsCmd.AddCommand(currentCmd)

	latestSandboxConfigCmd := &cobra.Command{
		Use:   "sandbox-config",
		Short: "View sandbox config",
		Long:  "View apps latest sandbox config",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
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
			svc := apps.New(c.v, c.apiClient, c.cfg)
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
			svc := apps.New(c.v, c.apiClient, c.cfg)
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
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.GetRunnerConfig(cmd.Context(), appID, PrintJSON)
		}),
	}
	latestRunnerConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestRunnerConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestRunnerConfig)

	selectAppCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current app",
		Long:  "Select your current app from a list or by app ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), appID, PrintJSON)
		}),
	}
	selectAppCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	appsCmd.AddCommand(selectAppCmd)

	unsetCurrentAppCmd := &cobra.Command{
		Use:   "unset-current",
		Short: "Unset your current app",
		Long:  "Unset your current app.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.UnsetCurrent(cmd.Context())
		}),
	}
	appsCmd.AddCommand(unsetCurrentAppCmd)

	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync nuon app directory",
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

			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.SyncDir(cmd.Context(), dirName, version.Version)
		}),
	}
	appsCmd.AddCommand(syncCmd)

	syncDirCmd := &cobra.Command{
		Deprecated:        "use `nuon sync` instead",
		Use:               "sync-dir",
		Short:             "Sync nuon app directory (deprecated)",
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

			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.DeprecatedSyncDir(cmd.Context(), dirName, version.Version)
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

			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.ValidateDir(cmd.Context(), dirName)
		}),
	}
	appsCmd.AddCommand(validateCmd)

	var (
		name string
	)
	createCmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a new app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), name, PrintJSON, noSelect)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "app name")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "do not automatically set the new app as the current app")

	appsCmd.AddCommand(createCmd)

	var confirmDelete bool
	deleteCmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete an existing app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Delete(cmd.Context(), appID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the app")
	deleteCmd.MarkFlagRequired("app-id")
	deleteCmd.MarkFlagRequired("confirm")

	appsCmd.AddCommand(deleteCmd)

	var rename bool
	renameCmd := &cobra.Command{
		Use:               "rename",
		Short:             "Rename an app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
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

	return appsCmd
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
			svc := variables.New(c.apiClient, c.cfg)
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
			svc := variables.New(c.apiClient, c.cfg)
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
			svc := variables.New(c.apiClient, c.cfg)
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
