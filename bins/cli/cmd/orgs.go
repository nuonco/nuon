package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/services/orgs"
)

func (c *cli) orgsCmd() *cobra.Command {
	orgsCmd := &cobra.Command{
		Use:               "orgs",
		Short:             "Manage your organizations",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	c.addOrgsInfoCommands(orgsCmd)
	c.addOrgsVCSCommands(orgsCmd)
	c.addOrgsCRUDCommands(orgsCmd)
	c.addOrgsInviteCommands(orgsCmd)

	return orgsCmd
}

func (c *cli) addOrgsInfoCommands(orgsCmd *cobra.Command) {
	var offset, limit int

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get current org",
		Long:  "Get the org you are currently authenticated with",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Current(cmd.Context(), PrintJSON)
		}),
	}
	currentCmd.Hidden = true
	orgsCmd.AddCommand(currentCmd)

	apiTokenCmd := &cobra.Command{
		Use:   "api-token",
		Short: "Get api token",
		Long:  "Get api token that is active for current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.APIToken(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(apiTokenCmd)

	idCmd := &cobra.Command{
		Use:   "id",
		Short: "Get current org id",
		Long:  "Get id for current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ID(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(idCmd)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List orgs",
		Long:    "List all your orgs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	orgsCmd.AddCommand(listCmd)

	orgsCmd.AddCommand(&cobra.Command{
		Use:   "print-config",
		Short: "Print the current cli config",
		Long:  "Print the current cli config being used",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.PrintConfig(PrintJSON)
		}),
	})
}

func (c *cli) addOrgsVCSCommands(orgsCmd *cobra.Command) {
	var connectionID string
	var offset, limit int

	listVCSConnections := &cobra.Command{
		Use:   "list-vcs-connections",
		Short: "List VCS connections",
		Long:  "List all connected GitHub accounts",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.VCSConnections(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listVCSConnections.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listVCSConnections.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	orgsCmd.AddCommand(listVCSConnections)

	deleteVCSConnectionCmd := &cobra.Command{
		Use:   "delete-vcs-connection",
		Short: "Delete VCS Connection",
		Long:  "Delete a VCS connection from your Nuon org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.DeleteVCSConnection(cmd.Context(), connectionID, PrintJSON)
		}),
	}
	deleteVCSConnectionCmd.MarkFlagRequired("org-id")
	deleteVCSConnectionCmd.Flags().StringVar(&connectionID, "connection-id", "", "The VCS Connection ID you want to remove")
	deleteVCSConnectionCmd.MarkFlagRequired("connection-id")
	orgsCmd.AddCommand(deleteVCSConnectionCmd)

	connectGithubCmd := &cobra.Command{
		Use:   "connect-github",
		Short: "Connect GitHub account",
		Long:  "Connect GitHub account to your Nuon org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ConnectGithub(cmd.Context())
		}),
	}
	connectGithubCmd.MarkFlagRequired("org-id")
	orgsCmd.AddCommand(connectGithubCmd)
}

func (c *cli) addOrgsCRUDCommands(orgsCmd *cobra.Command) {
	var id, name string
	var sandbox, noSelect bool

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create new org",
		Long:  "Create a new org and set it as the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), name, sandbox, noSelect, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name of your new org")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().BoolVar(&sandbox, "sandbox-mode", false, "Create org in sandbox mode")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "Do not automatically set the new org as the current org")
	orgsCmd.AddCommand(createCmd)

	selectOrgCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current org",
		Long:  "Select your current org from a list or by org ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), id, PrintJSON)
		}),
	}
	selectOrgCmd.Flags().StringVar(&id, "org", "", "The ID of the org you want to use")
	orgsCmd.AddCommand(selectOrgCmd)

	deselectOrgCmd := &cobra.Command{
		Use:   "deselect",
		Short: "Deselect your current org",
		Long:  "Deselect your current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Deselect(cmd.Context())
		}),
	}
	orgsCmd.AddCommand(deselectOrgCmd)
}

func (c *cli) addOrgsInviteCommands(orgsCmd *cobra.Command) {
	var email string
	var offset, limit int

	createInviteCmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a user to org",
		Long:  "Invite a user by email to org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.CreateInvite(cmd.Context(), email, PrintJSON)
		}),
	}
	createInviteCmd.Flags().StringVarP(&email, "email", "e", "", "Email of user to invite")
	orgsCmd.AddCommand(createInviteCmd)

	listInvitesCmd := &cobra.Command{
		Use:   "list-invites",
		Short: "List all org invites",
		Long:  "List all org invites",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ListInvites(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listInvitesCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listInvitesCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum invites to return")
	orgsCmd.AddCommand(listInvitesCmd)
}
