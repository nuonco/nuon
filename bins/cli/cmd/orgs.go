package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/spf13/cobra"
)

func newOrgsCmd(bindConfig config.BindCobraFunc, orgsService *orgs.Service) *cobra.Command {
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
		Run: func(cmd *cobra.Command, args []string) {
			orgsService.Current(cmd.Context(), PrintJSON)
		},
	}
	orgsCmd.AddCommand(currentCmd)

	return orgsCmd
}
