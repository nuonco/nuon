package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/spf13/cobra"
)

func (c *cli) orgsCmd() *cobra.Command {
	orgsCmd := &cobra.Command{
		Use:               "orgs",
		Short:             "Manage your organizations",
		PersistentPreRunE: c.persistentPreRunE,
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get current org",
		Long:  "Get the org you are currently authenticated with",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := orgs.New(c.apiClient)
			svc.Current(cmd.Context(), PrintJSON)
		},
	}
	orgsCmd.AddCommand(currentCmd)

	return orgsCmd
}
