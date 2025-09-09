package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/docs"
)

func (c *cli) docsCmd() *cobra.Command {
	docsCmd := &cobra.Command{
		Use:               "docs",
		Short:             "Explore the docs",
		Aliases:           []string{"d"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           HelpGroup.ID,
	}

	browseCmd := &cobra.Command{
		Use:   "browse",
		Short: "Browse documentation",
		Long:  "Open up documentation at https://docs.nuon.co",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := docs.New(c.cfg)
			return svc.Browse(cmd.Context(), PrintJSON)
		}),
	}
	docsCmd.AddCommand(browseCmd)

	apiBrowseCmd := &cobra.Command{
		Use:   "api",
		Short: "Open api-explorer",
		Long:  "Open api explorer with preauthorized key and org-id.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := docs.New(c.cfg)
			return svc.BrowseAPI(cmd.Context(), PrintJSON)
		}),
	}
	docsCmd.AddCommand(apiBrowseCmd)

	return docsCmd
}
