package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerApps(ctx context.Context) cobra.Command {

	appsCmd := &cobra.Command{
		Use:   "apps",
		Short: "View the apps in your org",
	}

	appsCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all your apps",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	appsCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the current app",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := c.api.GetApp(ctx, c.cfg.APP_ID)
			if err != nil {
				return err
			}

			ui.Line(ctx, "%s - %s", app.ID, app.Name)
			return nil
		},
	})

	return *appsCmd
}
