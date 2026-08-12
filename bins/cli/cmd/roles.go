package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/roles"
)

func (c *cli) rolesCmd() *cobra.Command {
	rolesCmd := &cobra.Command{
		Use:               "roles",
		Short:             "View and manage roles assignable to members and service accounts",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List roles assignable to members and service accounts",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.ListRoles(cmd.Context(), PrintJSON)
		}),
	}
	rolesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get <role>",
		Short: "Get a role and its permissions",
		Args:  cobra.ExactArgs(1),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.GetRole(cmd.Context(), args[0], PrintJSON)
		}),
	}
	rolesCmd.AddCommand(getCmd)

	var (
		title       string
		description string
		contexts    []string
		permissions []string
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom role",
		Long: `Create a custom role whose permissions are scoped to specific resources.

Custom roles are assigned like managed roles, addressed by role id:

  nuon orgs api-token create --role <role_id>`,
		Example: `  nuon roles create --title "app_web release manager" \
    --permission "read:app:app_web" \
    --permission "all:app_branch:*:scope=app_web"`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.CreateRole(cmd.Context(), title, description, contexts, permissions, PrintJSON)
		}),
	}
	createCmd.Flags().StringVar(&title, "title", "", "The role's display name (required)")
	createCmd.Flags().StringVar(&description, "description", "", "What the role grants")
	createCmd.Flags().StringSliceVar(&contexts, "context", nil, "Assignment surface the role may be offered on: team, service_account, api_token, oidc_trust_policy (repeatable)")
	createCmd.Flags().StringArrayVar(&permissions, "permission", nil, roles.PermissionFlagUsage)
	_ = createCmd.MarkFlagRequired("title")
	_ = createCmd.MarkFlagRequired("permission")
	rolesCmd.AddCommand(createCmd)

	var (
		editTitle       string
		editDescription string
		editContexts    []string
		editPermissions []string
	)

	editCmd := &cobra.Command{
		Use:   "edit <role_id>",
		Short: "Edit a custom role",
		Long: `Edit a custom role's metadata or replace its permissions. Managed roles cannot be edited.

Passing --permission replaces every existing entry, so pass the full set. Run
"nuon roles get <role_id>" to read the current entries in the same grammar.`,
		Args: cobra.ExactArgs(1),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.EditRole(cmd.Context(), args[0], editTitle, editDescription, editContexts, editPermissions, PrintJSON)
		}),
	}
	editCmd.Flags().StringVar(&editTitle, "title", "", "The role's display name")
	editCmd.Flags().StringVar(&editDescription, "description", "", "What the role grants")
	editCmd.Flags().StringSliceVar(&editContexts, "context", nil, "Replace the assignment surfaces the role may be offered on (repeatable)")
	editCmd.Flags().StringArrayVar(&editPermissions, "permission", nil, roles.PermissionFlagUsage)
	rolesCmd.AddCommand(editCmd)

	deleteCmd := &cobra.Command{
		Use:     "delete <role_id>",
		Aliases: []string{"rm"},
		Short:   "Delete a custom role",
		Args:    cobra.ExactArgs(1),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.DeleteRole(cmd.Context(), args[0], PrintJSON)
		}),
	}
	rolesCmd.AddCommand(deleteCmd)

	return rolesCmd
}
