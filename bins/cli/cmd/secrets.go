package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/secrets"
)

func (c *cli) secretsCmd() *cobra.Command {
	var (
		appID    string
		secretID string
		offset   int
		limit    int
	)

	secretsCmd := &cobra.Command{
		Deprecated:        "use 'nuon apps variables' instead",
		Use:               "secrets",
		Short:             "Create and manage secrets (deprecated)",
		Aliases:           []string{"s"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	// list command
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all secrets ",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := secrets.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to delete the secret from")
	listCmd.MarkFlagRequired("app-id")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "The offset to start listing secrets from")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "The number of secrets to list")
	secretsCmd.AddCommand(listCmd)

	// delete command
	confirmDelete := false
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete secret",
		Long:  "Delete an app secret",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := secrets.New(c.apiClient, c.cfg)
			return svc.Delete(cmd.Context(), appID, secretID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to delete the secret from")
	deleteCmd.Flags().StringVarP(&secretID, "secret-id", "s", "", "The ID or name of the secret to delete")
	deleteCmd.Flags().BoolVar(&confirmDelete, "confirm", false, "Confirm you want to delete the install")

	deleteCmd.MarkFlagRequired("app-id")
	deleteCmd.MarkFlagRequired("secret-id")
	deleteCmd.MarkFlagRequired("confirm")
	secretsCmd.AddCommand(deleteCmd)

	// create command
	var (
		name  string
		value string
	)
	createCmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a new secret.",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := secrets.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), appID, name, value, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name of the secret, must be alphanumeric, lower case and can use underscores.")
	createCmd.Flags().StringVarP(&value, "value", "v", "", "The secret value.")
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to delete the secret from")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("value")
	createCmd.MarkFlagRequired("app-id")
	secretsCmd.AddCommand(createCmd)

	return secretsCmd
}
