package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *cli) cloudConnectionsCmd() *cobra.Command {
	command := &cobra.Command{
		Use: "cloud-connections", Short: "Manage cloud connections",
		GroupID: AdditionalGroup.ID, PersistentPreRunE: c.persistentPreRunE,
	}
	outputs := outputsAnnotation(OutputTable, OutputJSON, OutputAgent)

	command.AddCommand(&cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List cloud connections", Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error { return c.cloudConnections.List(cmd.Context(), PrintJSON) }),
	})
	command.AddCommand(&cobra.Command{
		Use: "get <connection-id>", Short: "Get a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Get(cmd.Context(), args[0], PrintJSON)
		}),
	})

	var name, platform, targetID, principal, defaultRegion string
	var tenantID, subscriptionID, clientID, registry string
	var projectID, serviceAccountEmail, identityProvider string
	var capabilities, repositories []string
	create := &cobra.Command{
		Use: "create", Short: "Create a cloud connection", Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if subscriptionID != "" {
				targetID = subscriptionID
			}
			if clientID != "" {
				principal = clientID
			}
			if projectID != "" {
				targetID = projectID
			}
			if serviceAccountEmail != "" {
				principal = serviceAccountEmail
			}
			if targetID == "" {
				return fmt.Errorf("--target-id, --subscription-id, or --project-id is required")
			}
			if principal == "" {
				return fmt.Errorf("--principal, --client-id, or --service-account-email is required")
			}
			return c.cloudConnections.Create(cmd.Context(), name, platform, targetID, principal, tenantID, identityProvider, defaultRegion, registry, capabilities, repositories, PrintJSON)
		}),
	}
	create.Flags().StringVar(&name, "name", "", "Connection name")
	create.Flags().StringVar(&platform, "platform", "aws", "Cloud platform (aws, azure, or gcp)")
	create.Flags().StringVar(&targetID, "target-id", "", "Cloud account, subscription, or project ID")
	create.Flags().StringVar(&principal, "principal", "", "Cloud principal")
	create.Flags().StringVar(&tenantID, "tenant-id", "", "Azure Entra tenant ID")
	create.Flags().StringVar(&subscriptionID, "subscription-id", "", "Azure subscription ID")
	create.Flags().StringVar(&clientID, "client-id", "", "Azure application client ID")
	create.Flags().StringVar(&projectID, "project-id", "", "GCP project ID")
	create.Flags().StringVar(&serviceAccountEmail, "service-account-email", "", "GCP service account email")
	create.Flags().StringVar(&identityProvider, "identity-provider", "", "GCP Workload Identity Provider resource name")
	create.Flags().StringVar(&registry, "registry", "", "Azure Container Registry name or login server")
	create.Flags().StringVar(&defaultRegion, "default-region", "", "Default AWS region")
	create.Flags().StringSliceVar(&capabilities, "capability", nil, "Capability: stacks or images (repeatable)")
	create.Flags().StringSliceVar(&repositories, "repository", nil, "Container repository name (repeatable)")
	_ = create.MarkFlagRequired("name")
	_ = create.MarkFlagRequired("capability")
	command.AddCommand(create)

	var verifyRepositories []string
	var verifyRegistry string
	verify := &cobra.Command{
		Use: "verify <connection-id>", Short: "Verify a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Verify(cmd.Context(), args[0], verifyRegistry, verifyRepositories, PrintJSON)
		}),
	}
	verify.Flags().StringVar(&verifyRegistry, "registry", "", "Azure Container Registry name or login server")
	verify.Flags().StringSliceVar(&verifyRepositories, "repository", nil, "Container repository name to probe (repeatable)")
	command.AddCommand(verify)

	command.AddCommand(&cobra.Command{
		Use: "delete <connection-id>", Short: "Delete an unused cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Delete(cmd.Context(), args[0], PrintJSON)
		}),
	})
	return command
}
