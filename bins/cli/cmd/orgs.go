package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/spf13/cobra"
)

func registerOrgs(orgsService *orgs.Service) cobra.Command {
	orgsCmd := &cobra.Command{
		Use:   "orgs",
		Short: "Manage your organizations",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get current org",
		Long:  "Get the org you are currently authenticated with",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return orgsService.Current(cmd.Context())
		},
	}
	orgsCmd.AddCommand(currentCmd)

	return *orgsCmd
}
