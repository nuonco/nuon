package cmd

import (
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerApps() cobra.Command {

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
			ctx := cmd.Context()
			apps, err := c.api.GetApps(ctx)
			if err != nil {
				return err
			}

			for _, app := range apps {
				ui.Line(ctx, "%s - %s", app.ID, app.Name)
			}

			return nil
		},
	})

	appID := ""
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get the current app",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			app, err := c.api.GetApp(ctx, appID)
			if err != nil {
				return err
			}

			ui.Line(ctx, "%s - %s", app.ID, app.Name)
			return nil
		},
	}
	// TODO: Update API to support getting app by name and add a flag for that.
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app")
	getCmd.MarkFlagRequired("app-id")
	appsCmd.AddCommand(getCmd)

	return *appsCmd
}
