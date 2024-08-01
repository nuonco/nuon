package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
)

func (c *cli) orgsCmd() *cobra.Command {
	var (
		id      string
		name    string
		sandbox bool
		limit   int64
		email   string
	)

	orgsCmd := &cobra.Command{
		Use:               "orgs",
		Short:             "Manage your organizations",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get current org",
		Long:  "Get the org you are currently authenticated with",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Current(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(currentCmd)

	healthChecksCmd := &cobra.Command{
		Use:   "health-checks",
		Short: "List health checks",
		Long:  "List recent helath checks for the org you are currently authenticated with",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ListHealthChecks(cmd.Context(), limit, PrintJSON)
		}),
	}
	healthChecksCmd.Flags().Int64VarP(&limit, "limit", "l", 60, "Maximum health checks to return")
	orgsCmd.AddCommand(healthChecksCmd)

	apiTokenCmd := &cobra.Command{
		Use:   "api-token",
		Short: "Get api token",
		Long:  "Get api token that is active for current org",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.APIToken(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(apiTokenCmd)

	idCmd := &cobra.Command{
		Use:   "id",
		Short: "Get current org id",
		Long:  "Get id for current org",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
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
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(listCmd)

	listConntectedRepos := &cobra.Command{
		Use:   "list-connected-repos",
		Short: "List connected repos",
		Long:  "List repositories from connected GitHub accounts",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ConnectedRepos(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(listConntectedRepos)

	listVCSConnections := &cobra.Command{
		Use:   "list-vcs-connections",
		Short: "List VCS connections",
		Long:  "List all connected GitHub accounts",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.VCSConnections(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(listVCSConnections)

	connectGithubCmd := &cobra.Command{
		Use:   "connect-github",
		Short: "Connect GitHub account",
		Long:  "Connect GitHub account to your Nuon org",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ConnectGithub(cmd.Context())
		}),
	}
	connectGithubCmd.Flags().StringVarP(&id, "org-id", "o", "", "The ID of the org you want to use")
	connectGithubCmd.MarkFlagRequired("org-id")
	orgsCmd.AddCommand(connectGithubCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create new org",
		Long:  "Create a new org and set it as the current org",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), name, sandbox, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name of your new org")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().BoolVar(&sandbox, "sandbox-mode", false, "Create org in sandbox mode")
	orgsCmd.AddCommand(createCmd)

	selectOrgCmd := &cobra.Command{
		Use:   "select",
		Short: "Select your current org",
		Long:  "Select your current org from a list or by org ID",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), id, PrintJSON)
		}),
	}
	selectOrgCmd.Flags().StringVar(&id, "org", "", "The ID of the org you want to use")
	orgsCmd.AddCommand(selectOrgCmd)

	orgsCmd.AddCommand(&cobra.Command{
		Use:   "print-config",
		Short: "Print the current cli config",
		Long:  "Print the current cli config being used",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.PrintConfig(PrintJSON)
		}),
	})

	createInviteCmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a user to org",
		Long:  "Invite a user by email to org",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
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
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ListInvites(cmd.Context(), limit, PrintJSON)
		}),
	}
	listInvitesCmd.Flags().Int64VarP(&limit, "limit", "l", 5, "Maximum invites to return")
	orgsCmd.AddCommand(listInvitesCmd)

	return orgsCmd
}
