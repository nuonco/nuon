package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
)

func (c *cli) appsCmd() *cobra.Command {
	var (
		all  bool
		file string
	)

	appsCmd := &cobra.Command{
		Use:               "apps",
		Short:             "View the apps in your org",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.List(cmd.Context(), PrintJSON)
		},
	}

	appsCmd.AddCommand(listCmd)

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an app",
		Long:  "Get either the current app or an app by name or ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Get(cmd.Context(), appID, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get the current app",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Get(cmd.Context(), c.cfg.GetString("app_id"), PrintJSON)
		},
	}
	appsCmd.AddCommand(currentCmd)

	latestSandboxConfigCmd := &cobra.Command{
		Use:   "sandbox-config",
		Short: "View sandbox config",
		Long:  "View apps latest sandbox config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.GetSandboxConfig(cmd.Context(), appID, PrintJSON)
		},
	}
	latestSandboxConfigCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestSandboxConfigCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestSandboxConfigCmd)

	configs := &cobra.Command{
		Use:   "configs",
		Short: "List app configs",
		Long:  "List app configs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.ListConfigs(cmd.Context(), appID, PrintJSON)
		},
	}
	configs.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	configs.MarkFlagRequired("app-id")
	appsCmd.AddCommand(configs)

	exportTerraform := &cobra.Command{
		Use:   "export-terraform",
		Short: "Export terraform config",
		Long:  "Export terraform config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.ExportTerraform(cmd.Context(), appID, PrintJSON)
		},
	}
	exportTerraform.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	exportTerraform.MarkFlagRequired("app-id")
	appsCmd.AddCommand(exportTerraform)

	latestInputConfig := &cobra.Command{
		Use:   "input-config",
		Short: "View app input config",
		Long:  "View latest app input config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.GetInputConfig(cmd.Context(), appID, PrintJSON)
		},
	}
	latestInputConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestInputConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestInputConfig)

	latestRunnerConfig := &cobra.Command{
		Use:   "runner-config",
		Short: "View app runner config",
		Long:  "View latest app runner config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.GetRunnerConfig(cmd.Context(), appID, PrintJSON)
		},
	}
	latestRunnerConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestRunnerConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestRunnerConfig)

	selectAppCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current app",
		Long:  "Select your current app from a list or by app ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Select(cmd.Context(), appID, PrintJSON)
		},
	}
	selectAppCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	appsCmd.AddCommand(selectAppCmd)

	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync all .nuon.toml config files in the current directory.",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Sync(cmd.Context(), all, file, PrintJSON)
		},
	}
	syncCmd.Flags().StringVarP(&file, "file", "c", "", "Config file to sync")
	syncCmd.Flags().BoolVarP(&all, "all", "", false, "sync all config files found")
	appsCmd.AddCommand(syncCmd)

	validateCmd := &cobra.Command{
		Use:               "validate",
		Short:             "Validate a toml config file",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Validate(cmd.Context(), all, file, PrintJSON)
		},
	}
	validateCmd.Flags().StringVarP(&file, "file", "c", "", "Config file to sync")
	validateCmd.Flags().BoolVarP(&all, "all", "", false, "sync all config files found")
	appsCmd.AddCommand(validateCmd)

	var (
		name     string
		template string
	)
	createCmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a new app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Create(cmd.Context(), name, template, PrintJSON)
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "app name")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&template, "template", "", "aws-ecs", "app config template type")
	appsCmd.AddCommand(createCmd)

	var confirmDelete bool
	deleteCmd := &cobra.Command{
		Use:               "delete",
		Short:             "Delete an existing app",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Delete(cmd.Context(), appID, PrintJSON)
		},
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
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			svc.Rename(cmd.Context(), appID, name, rename, PrintJSON)
		},
	}
	renameCmd.Flags().StringVarP(&name, "name", "n", "", "app name")
	renameCmd.MarkFlagRequired("name")
	renameCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	renameCmd.MarkFlagRequired("app-id")
	renameCmd.Flags().BoolVarP(&rename, "rename", "", true, "Rename config file if it exists")

	appsCmd.AddCommand(renameCmd)

	return appsCmd
}
