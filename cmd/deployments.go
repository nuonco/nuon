package cmd

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/nuonctl/internal"
	temporalclient "github.com/powertoolsdev/nuonctl/internal/clients/temporal"
	"github.com/powertoolsdev/nuonctl/internal/commands/deployments"
	"github.com/powertoolsdev/nuonctl/internal/repos/executors"
	"github.com/powertoolsdev/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/nuonctl/internal/repos/workflows"
	"github.com/spf13/cobra"
)

func registerDeployments(ctx context.Context, v *validator.Validate, rootCmd *cobra.Command) error {
	// load configuration and setup the deployments command namespace
	var cfg internal.Config
	if err := config.LoadInto(rootCmd.Flags(), &cfg); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}
	if err := cfg.Validate(v); err != nil {
		return fmt.Errorf("unable to validate config: %w", err)
	}

	// TODO(jm): move this into a pre-run, so it isn't executed before help commands
	tclient, err := temporalclient.New(v, temporalclient.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	workflows, err := workflows.New(v, workflows.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	temporal, err := temporal.New(v, temporal.WithClient(tclient))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	executors, err := executors.New(v, executors.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	deploys, err := deployments.New(v,
		deployments.WithTemporalRepo(temporal),
		deployments.WithWorkflowsRepo(workflows),
		deployments.WithExecutorsRepo(executors))
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

	return nil
}
