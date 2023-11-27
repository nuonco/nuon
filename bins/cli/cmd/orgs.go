package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
	"github.com/spf13/cobra"
)

func (c *cli) orgsCmd() *cobra.Command {
	var id string

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

	setCurrentCmd := &cobra.Command{
		Use:   "set-current",
		Short: "Set current org",
		Long:  "Set current org by org ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := orgs.New(c.apiClient)
			svc.SetCurrent(cmd.Context(), id, c.cfg)
		},
	}
	setCurrentCmd.Flags().StringVarP(&id, "org-id", "o", "", "The ID of the org you want to use")
	setCurrentCmd.MarkFlagRequired("org-id")
	orgsCmd.AddCommand(setCurrentCmd)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List orgs",
		Long:    "List all your orgs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := orgs.New(c.apiClient)
			svc.List(cmd.Context(), PrintJSON)
		},
	}
	orgsCmd.AddCommand(listCmd)

	listConntectedRepos := &cobra.Command{
		Use:   "list-connected-repos",
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
