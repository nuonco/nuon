package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/installs"
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
		Use:     "installs",
		Short:   "commands for working with installs",
		Aliases: []string{"i"},
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

	installsCmd.AddCommand(&cobra.Command{
		Use:   "print-request",
		Short: "print any request by passing in a key - works for instances as well",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.PrintRequest(ctx, args[0])
		},
	})
	installsCmd.AddCommand(&cobra.Command{
		Use:   "print-response",
		Short: "print any response by passing in a key - works for instances as well",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.PrintResponse(ctx, args[0])
		},
	})
	installsCmd.AddCommand(&cobra.Command{
		Use:   "reprovision",
		Short: "reprovision an install by re-triggering workflow with it's original request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.Reprovision(ctx, installID)
		},
	})
	installsCmd.AddCommand(&cobra.Command{
		Use:   "deprovision",
		Short: "deprovision an install by submitting deprovision with it's original request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.Deprovision(ctx, installID)
		},
	})

	installsCmd.AddCommand(&cobra.Command{
		Use:   "deprovision-bulk",
		Short: "deprovision all installs that are passed in",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installs.DeprovisionBulk(ctx, args)
		},
	})
	return nil
}
