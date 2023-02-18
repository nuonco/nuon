package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/commands/installs"
	"github.com/spf13/cobra"
)

func (c *cli) registerInstalls(ctx context.Context, rootCmd *cobra.Command) error {
	installs, err := installs.New(c.v,
		installs.WithTemporalRepo(c.temporal),
		installs.WithWorkflowsRepo(c.workflows),
		installs.WithExecutorsRepo(c.executors),
	)
	if err != nil {
		return fmt.Errorf("unable to get initialize deploys: %w", err)
	}

	// register commands
	var installsCmd = &cobra.Command{
		Use:   "installs",
		Short: "commands for working with installs",
	}
	rootCmd.AddCommand(installsCmd)

	var installID string
	installsCmd.PersistentFlags().StringVar(&installID, "install-id", "", "install short id")

	installsCmd.AddCommand(&cobra.Command{
		Use: "print-provision-request",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.PrintProvisionRequest(ctx, installID)
		},
	})
	return nil
}
