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
			svc := apps.New(c.apiClient)
			svc.List(cmd.Context(), PrintJSON)
		},
	}

	appsCmd.AddCommand(listCmd)

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get the current app",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient)
			svc.Get(cmd.Context(), appID, PrintJSON)
		},
	}
	// TODO: Update API to support getting app by name and add a flag for that.
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	return appsCmd
}
