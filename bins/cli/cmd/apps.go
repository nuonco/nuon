package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/spf13/cobra"
)

func registerApps(appsService *apps.Service) cobra.Command {

	appsCmd := &cobra.Command{
		Use:   "apps",
		Short: "View the apps in your org",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	appsCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return appsService.List(cmd.Context())
		},
	})

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get the current app",
		RunE: func(cmd *cobra.Command, args []string) error {
			return appsService.Get(cmd.Context(), appID)
		},
	}
	// TODO: Update API to support getting app by name and add a flag for that.
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	return *appsCmd
}
