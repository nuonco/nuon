package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerContext(ctx context.Context) cobra.Command {
	contextCmd := &cobra.Command{
		Use:   "context",
		Short: "Get current org context",
		Long:  "Get the current org context",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			org, err := c.api.GetOrg(ctx)
			if err != nil {
				return err
			}

			statusColor := ui.GetStatusColor(org.Status)
			ui.Line(ctx, "%s%s %s- %s - %s", statusColor, org.Status, ui.ColorReset, org.ID, org.Name)
			return nil
		},
	}

	return *contextCmd
}
