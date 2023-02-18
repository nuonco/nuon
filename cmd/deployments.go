package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/commands/deployments"
	"github.com/spf13/cobra"
)

func (c *cli) registerDeployments(ctx context.Context, rootCmd *cobra.Command) error {
	deploys, err := deployments.New(c.v,
		deployments.WithTemporalRepo(c.temporal),
		deployments.WithWorkflowsRepo(c.workflows),
		deployments.WithExecutorsRepo(c.executors))
	if err != nil {
		return fmt.Errorf("unable to get initialize deploys: %w", err)
	}

	// register commands
	var deploymentsCmd = &cobra.Command{
		Use:   "deployments",
		Short: "commands for working with deployments",
	}
	rootCmd.AddCommand(deploymentsCmd)

	var installID string
	deploymentsCmd.PersistentFlags().StringVar(&installID, "install-id", "", "install short id")
	var preset string
	deploymentsCmd.PersistentFlags().StringVar(&preset, "preset", "", "component preset name")

	deploymentsCmd.AddCommand(&cobra.Command{
		Use: "install-preset",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.TriggerInstallPreset(ctx, installID, preset)
		},
	})

	deploymentsCmd.AddCommand(&cobra.Command{
		Use: "install-build-plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.InstallBuildPlan(ctx, installID, preset)
		},
	})
	deploymentsCmd.AddCommand(&cobra.Command{
		Use: "install-sync-image-plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.InstallSyncImagePlan(ctx, installID, preset)
		},
	})
	deploymentsCmd.AddCommand(&cobra.Command{
		Use: "install-deploy-plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.InstallDeployPlan(ctx, installID, preset)
		},
	})

	deploymentsCmd.AddCommand(&cobra.Command{
		Use: "print-preset",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.PrintPreset(ctx, preset)
		},
	})

	deploymentsCmd.AddCommand(&cobra.Command{
		Use:   "print-request",
		Short: "print any request by passing in a key - works for instances as well",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.PrintPreset(ctx, args[0])
		},
	})
	deploymentsCmd.AddCommand(&cobra.Command{
		Use:   "print-response",
		Short: "print any response by passing in a key - works for instances as well",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.PrintResponse(ctx, args[0])
		},
	})
	deploymentsCmd.AddCommand(&cobra.Command{
		Use:   "print-plan",
		Short: "print any plan by passing in a key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploys.PrintPlan(ctx, args[0])
		},
	})

	return nil
}
