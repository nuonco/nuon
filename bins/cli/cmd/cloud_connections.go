package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *cli) cloudConnectionsCmd() *cobra.Command {
	command := &cobra.Command{
		Use: "cloud-connections", Short: "Manage cloud connections",
		GroupID: AdditionalGroup.ID, PersistentPreRunE: c.persistentPreRunE,
	}
	outputs := outputsAnnotation(OutputTable, OutputJSON, OutputAgent)

	var offset, limit int
	list := &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List cloud connections", Args: cobra.NoArgs, Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.cloudConnections.List(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	list.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	list.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum cloud connections to return")
	command.AddCommand(list)
	command.AddCommand(&cobra.Command{
		Use: "get <connection-id>", Short: "Get a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Get(cmd.Context(), args[0], PrintJSON)
		}),
	})

	var name, platform, targetID, principal, defaultRegion, preset, projectNumber string
	var azureTenantID, azureRegistryID, azureLoginServer string
	var repositoryARNs, garRepositories, acrRepositories []string
	create := &cobra.Command{
		Use: "create", Short: "Create a cloud connection", Args: cobra.NoArgs, Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if targetID == "" {
				return ui.PrintError(fmt.Errorf("--target-id is required"))
			}
			if principal == "" {
				return ui.PrintError(fmt.Errorf("--principal is required"))
			}
			if (preset == "ecr" && len(repositoryARNs) == 0 || preset == "gar" && len(garRepositories) == 0) && defaultRegion == "" {
				return ui.PrintError(fmt.Errorf("--default-region is required for image connections without initial repository setup input"))
			}
			if preset == "gar" && projectNumber == "" {
				return ui.PrintError(fmt.Errorf("--gcp-project-number is required with --preset gar"))
			}
			if preset == "acr" && (azureTenantID == "" || azureRegistryID == "" || azureLoginServer == "") {
				return ui.PrintError(fmt.Errorf("--azure-tenant-id, --azure-registry-id, and --azure-registry-login-server are required with --preset acr"))
			}
			return c.cloudConnections.Create(cmd.Context(), &models.ServiceCreateRequest{
				Name: name, Platform: platform, TargetID: targetID, Principal: principal,
				DefaultRegion: defaultRegion, Preset: models.AppCloudConnectionPreset(preset),
				EcrRepositoryArns: repositoryARNs, GcpProjectNumber: projectNumber, GarRepositories: garRepositories,
				AzureTenantID: azureTenantID, AzureRegistryID: azureRegistryID,
				AzureRegistryLoginServer: azureLoginServer, AcrRepositories: acrRepositories,
			}, PrintJSON)
		}),
	}
	create.Flags().StringVar(&name, "name", "", "Connection name")
	create.Flags().StringVar(&platform, "platform", "aws", "Cloud platform: aws, gcp, or azure")
	create.Flags().StringVar(&targetID, "target-id", "", "AWS account ID, GCP project ID, or Azure subscription ID")
	create.Flags().StringVar(&principal, "principal", "", "AWS IAM role ARN, GCP service account email, or Entra application client ID")
	create.Flags().StringVar(&defaultRegion, "default-region", "", "AWS region or GAR location for image connection matching (inferred from initial repository input when provided)")
	create.Flags().StringVar(&preset, "preset", "", "Access preset: stacks, ecr, gar, acr (image repository access), or custom")
	create.Flags().StringSliceVar(&repositoryARNs, "repository-arn", nil, "Optional ECR repository ARN for initial permission commands (repeatable; not saved)")
	create.Flags().StringVar(&projectNumber, "gcp-project-number", "", "Numeric GCP project number (required with --preset gar)")
	create.Flags().StringSliceVar(&garRepositories, "gar-repository", nil, "Optional GAR resource projects/PROJECT/locations/LOCATION/repositories/REPOSITORY for initial permission commands (repeatable; not saved)")
	create.Flags().StringVar(&azureTenantID, "azure-tenant-id", "", "Microsoft Entra tenant ID (required with --preset acr)")
	create.Flags().StringVar(&azureRegistryID, "azure-registry-id", "", "Full ARM resource ID of the existing ACR registry (required with --preset acr)")
	create.Flags().StringVar(&azureLoginServer, "azure-registry-login-server", "", "ACR login server, e.g. acme.azurecr.io (required with --preset acr)")
	create.Flags().StringSliceVar(&acrRepositories, "acr-repository", nil, "Optional ACR repository path, e.g. acme/api, for initial permission commands (repeatable; not saved)")
	_ = create.MarkFlagRequired("name")
	_ = create.MarkFlagRequired("preset")
	command.AddCommand(create)

	var wait, noWait bool
	verify := &cobra.Command{
		Use: "verify <connection-id>", Short: "Verify a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Verify(cmd.Context(), args[0], wait && !noWait, PrintJSON)
		}),
	}
	verify.Flags().BoolVar(&wait, "wait", true, "Wait up to two minutes for verification")
	verify.Flags().BoolVar(&noWait, "no-wait", false, "Enqueue verification without waiting")
	verify.MarkFlagsMutuallyExclusive("wait", "no-wait")
	command.AddCommand(verify)

	command.AddCommand(&cobra.Command{
		Use: "delete <connection-id>", Short: "Delete an unused cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Delete(cmd.Context(), args[0], PrintJSON)
		}),
	})
	return command
}
