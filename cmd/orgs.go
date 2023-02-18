package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/nuonctl/internal/commands/orgs"
	"github.com/spf13/cobra"
)

func (c *cli) registerOrgs(ctx context.Context, rootCmd *cobra.Command) error {
	orgs, err := orgs.New(c.v, orgs.WithTemporalRepo(c.temporal), orgs.WithWorkflowsRepo(c.workflows))
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

	orgsCmd.AddCommand(&cobra.Command{
		Use: "provision",

		// TODO(jm): do not use RunE and provide own formatter
		RunE: func(cmd *cobra.Command, args []string) error {
			return orgs.Provision(ctx, orgID)
		},
	})
	return nil
}
