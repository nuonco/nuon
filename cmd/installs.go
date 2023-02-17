package cmd

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/nuonctl/internal"
	temporalclient "github.com/powertoolsdev/nuonctl/internal/clients/temporal"
	"github.com/powertoolsdev/nuonctl/internal/commands/installs"
	"github.com/powertoolsdev/nuonctl/internal/repos/executors"
	"github.com/powertoolsdev/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/nuonctl/internal/repos/workflows"
	"github.com/spf13/cobra"
)

func registerInstalls(ctx context.Context, v *validator.Validate, rootCmd *cobra.Command) error {
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

	executors, err := executors.New(v, executors.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	temporal, err := temporal.New(v, temporal.WithClient(tclient))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	workflows, err := workflows.New(v, workflows.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	installs, err := installs.New(v,
		installs.WithTemporalRepo(temporal),
		installs.WithWorkflowsRepo(workflows),
		installs.WithExecutorsRepo(executors),
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
