package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *cli) imageRepositoriesCmd() *cobra.Command {
	command := &cobra.Command{
		Use: "image-repositories", Short: "Manage private image repositories and the cloud connections that pull them",
		GroupID: AdditionalGroup.ID, PersistentPreRunE: c.persistentPreRunE,
	}
	outputs := outputsAnnotation(OutputTable, OutputJSON, OutputAgent)

	var query, platform string
	var offset, limit int
	list := &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List image repositories", Args: cobra.NoArgs, Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.cloudConnections.ListRepositories(cmd.Context(), query, platform, offset, limit, PrintJSON)
		}),
	}
	list.Flags().StringVarP(&query, "query", "q", "", "Filter by image URL")
	list.Flags().StringVar(&platform, "platform", "", "Filter by cloud platform: aws, gcp, or azure")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	list.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum image repositories to return")
	command.AddCommand(list)

	command.AddCommand(&cobra.Command{
		Use: "get <repository-id>", Short: "Get an image repository and its permission commands", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.GetRepository(cmd.Context(), args[0], PrintJSON)
		}),
	})

	var repository, connectionID string
	create := &cobra.Command{
		Use: "create", Short: "Assign an image repository to a cloud connection", Args: cobra.NoArgs, Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.cloudConnections.CreateRepository(cmd.Context(), &models.ServiceCreateImageRepositoryRequest{
				Repository: &repository, CloudConnectionID: &connectionID,
			}, PrintJSON)
		}),
	}
	create.Flags().StringVar(&repository, "repository", "", "ECR repository ARN, GAR resource projects/PROJECT/locations/LOCATION/repositories/REPOSITORY, or ACR repository path such as acme/api")
	create.Flags().StringVar(&connectionID, "connection-id", "", "Image cloud connection that pulls this repository")
	_ = create.MarkFlagRequired("repository")
	_ = create.MarkFlagRequired("connection-id")
	command.AddCommand(create)

	command.AddCommand(&cobra.Command{
		Use: "delete <repository-id>", Short: "Remove an image repository assignment", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.DeleteRepository(cmd.Context(), args[0], PrintJSON)
		}),
	})
	return command
}
