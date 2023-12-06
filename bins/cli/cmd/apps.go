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
		Short: "Get the current app",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.Get(cmd.Context(), appID, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

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
			svc.SetCurrent(cmd.Context(), appID)
		},
	}
	setCurrentCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app")
	setCurrentCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(setCurrentCmd)

	return appsCmd
}
