package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/commands/apps"
	"github.com/spf13/cobra"
)

func (c *cli) registerApps(ctx context.Context, rootCmd *cobra.Command) error {
	apps, err := apps.New(c.v, apps.WithWorkflowsRepo(c.workflows))
	if err != nil {
		return fmt.Errorf("unable to get initialize apps: %w", err)
	}

	// register commands
	var appsCmd = &cobra.Command{
		Use:     "apps",
		Aliases: []string{"a"},
		Short:   "commands for working with apps",
	}
	rootCmd.AddCommand(appsCmd)

	var appID string
	appsCmd.PersistentFlags().StringVar(&appID, "app-id", "", "app short id")

	var orgID string
	appsCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "org short id")

	appsCmd.AddCommand(&cobra.Command{
		Use: "print-provision-request",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return apps.PrintProvisionRequest(ctx, orgID, appID)
		},
	})
	appsCmd.AddCommand(&cobra.Command{
		Use: "print-provision-response",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return apps.PrintProvisionResponse(ctx, orgID, appID)
		},
	})
	return nil
}
