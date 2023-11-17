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

	listConntectedRepos := &cobra.Command{
		Use:   "list-conntected-repos",
		Short: "List connected repos",
		Long:  "List repositories from connected GitHub accounts",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := orgs.New(c.apiClient)
			svc.ConnectedRepos(cmd.Context(), PrintJSON)
		},
	}
	orgsCmd.AddCommand(listConntectedRepos)

	return orgsCmd
}
