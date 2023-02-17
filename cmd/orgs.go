package cmd

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/nuonctl/internal"
	temporalclient "github.com/powertoolsdev/nuonctl/internal/clients/temporal"
	"github.com/powertoolsdev/nuonctl/internal/commands/orgs"
	"github.com/powertoolsdev/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/nuonctl/internal/repos/workflows"
	"github.com/spf13/cobra"
)

func registerOrgs(ctx context.Context, v *validator.Validate, rootCmd *cobra.Command) error {
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

	temporal, err := temporal.New(v, temporal.WithClient(tclient))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	workflows, err := workflows.New(v, workflows.WithConfig(&cfg))
	if err != nil {
		return fmt.Errorf("unable to get temporal client: %w", err)
	}

	orgs, err := orgs.New(v, orgs.WithTemporalRepo(temporal), orgs.WithWorkflowsRepo(workflows))
	if err != nil {
		return fmt.Errorf("unable to get initialize deploys: %w", err)
	}

	// register commands
	var orgsCmd = &cobra.Command{
		Use:   "orgs",
		Short: "commands for working with orgs",
	}
	rootCmd.AddCommand(orgsCmd)

	var orgID string
	orgsCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "org short id")

	orgsCmd.AddCommand(&cobra.Command{
		Use: "print-provision-request",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return orgs.PrintProvisionRequest(ctx, orgID)
		},
	})
	orgsCmd.AddCommand(&cobra.Command{
		Use: "print-provision-response",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return orgs.PrintProvisionResponse(ctx, orgID)
		},
	})
	return nil
}
