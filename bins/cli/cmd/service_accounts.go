package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/serviceaccounts"
)

func (c *cli) serviceAccountsCmd() *cobra.Command {
	var (
		id             string
		name           string
		role           string
		offset         int
		limit          int
		includeRunners bool
		duration       string
		invalidate     bool
	)

	serviceAccountsCmd := &cobra.Command{
		Use:               "service-accounts",
		Short:             "Manage machine users for the current org",
		Long:              "Create, list, update, and delete service accounts (machine users) and their API tokens for the current org",
		Aliases:           []string{"service-account", "sa"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List service accounts for the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.ListServiceAccounts(cmd.Context(), includeRunners, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "The offset of results to return")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "The limit of results to return")
	listCmd.Flags().BoolVar(&includeRunners, "include-runners", false, "Include service accounts with the runner role (excluded by default)")
	serviceAccountsCmd.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a service account for the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.CreateServiceAccount(cmd.Context(), name, role, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "A human-friendly name for the service account")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVar(&role, "role", "", "The role to grant the service account: a built-in value or a custom role id; see 'nuon roles list'")
	createCmd.MarkFlagRequired("role")
	serviceAccountsCmd.AddCommand(createCmd)

	updateNameCmd := &cobra.Command{
		Use:   "update-name",
		Short: "Update the name of a service account",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.UpdateServiceAccountName(cmd.Context(), id, name, PrintJSON)
		}),
	}
	updateNameCmd.Flags().StringVar(&id, "id", "", "The ID of the service account")
	updateNameCmd.MarkFlagRequired("id")
	updateNameCmd.Flags().StringVarP(&name, "name", "n", "", "The new name for the service account")
	updateNameCmd.MarkFlagRequired("name")
	serviceAccountsCmd.AddCommand(updateNameCmd)

	updateRoleCmd := &cobra.Command{
		Use:   "update-role",
		Short: "Update the role of a service account",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.UpdateServiceAccountRole(cmd.Context(), id, role, PrintJSON)
		}),
	}
	updateRoleCmd.Flags().StringVar(&id, "id", "", "The ID of the service account")
	updateRoleCmd.MarkFlagRequired("id")
	updateRoleCmd.Flags().StringVar(&role, "role", "", "The role to grant the service account: a built-in value or a custom role id; see 'nuon roles list'")
	updateRoleCmd.MarkFlagRequired("role")
	serviceAccountsCmd.AddCommand(updateRoleCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a service account for the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.DeleteServiceAccount(cmd.Context(), id, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVar(&id, "id", "", "The ID of the service account to delete")
	deleteCmd.MarkFlagRequired("id")
	serviceAccountsCmd.AddCommand(deleteCmd)

	tokensCmd := &cobra.Command{
		Use:               "tokens",
		Short:             "Manage API tokens for a service account",
		PersistentPreRunE: c.persistentPreRunE,
	}

	createTokenCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API token for a service account",
		Long:  "Create an API token for a service account. The token is only shown once.",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := serviceaccounts.New(c.apiClient, c.cfg)
			return svc.CreateServiceAccountToken(cmd.Context(), id, duration, invalidate, PrintJSON)
		}),
	}
	createTokenCmd.Flags().StringVar(&id, "id", "", "The ID of the service account")
	createTokenCmd.MarkFlagRequired("id")
	createTokenCmd.Flags().StringVar(&duration, "duration", "8760h", "How long the token is valid (Go duration, e.g. 720h)")
	createTokenCmd.Flags().BoolVar(&invalidate, "invalidate", false, "Invalidate the service account's existing tokens before creating the new one")
	tokensCmd.AddCommand(createTokenCmd)
	serviceAccountsCmd.AddCommand(tokensCmd)

	return serviceAccountsCmd
}
