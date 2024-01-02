package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/spf13/cobra"
)

func (c *cli) appsCmd() *cobra.Command {
	appsCmd := &cobra.Command{
		Use:               "apps",
		Short:             "View the apps in your org",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
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
			svc := apps.New(c.apiClient, c.cfg)
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
			svc := apps.New(c.apiClient, c.cfg)
			svc.Get(cmd.Context(), c.cfg.GetString("app_id"), PrintJSON)
		},
	}
	appsCmd.AddCommand(currentCmd)

	latestSandboxConfigCmd := &cobra.Command{
		Use:   "sandbox-config",
		Short: "View sandbox config",
		Long:  "View apps latest sandbox config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.GetSandboxConfig(cmd.Context(), appID, PrintJSON)
		},
	}
	latestSandboxConfigCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestSandboxConfigCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestSandboxConfigCmd)

	latestInputConfig := &cobra.Command{
		Use:   "input-config",
		Short: "View app input config",
		Long:  "View latest app input config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.GetInputConfig(cmd.Context(), appID, PrintJSON)
		},
	}
	latestInputConfig.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	latestInputConfig.MarkFlagRequired("app-id")
	appsCmd.AddCommand(latestInputConfig)

	setCurrentCmd := &cobra.Command{
		Use:   "set-current",
		Short: "Set current app",
		Long:  "Set current app by app ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.SetCurrent(cmd.Context(), appID, PrintJSON)
		},
	}
	setCurrentCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	setCurrentCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(setCurrentCmd)

	selectAppCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current app",
		Long:  "Select your current app from a list or by app ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.Select(cmd.Context(), appID, PrintJSON)
		},
	}
	selectAppCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	appsCmd.AddCommand(selectAppCmd)

	return appsCmd
}
