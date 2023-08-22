package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerInstalls(ctx context.Context, rootCmd *cobra.Command) error {
	var (
		id     string
		name   string
		arn    string
		region string
	)

	orgID := os.Getenv("NUON_ORG_ID")
	appID := os.Getenv("NUON_APP_ID")

	apiClient, err := client.New(c.v, client.WithAuthToken(os.Getenv("NUON_API_TOKEN")), client.WithURL(os.Getenv("NUON_API_URL")), client.WithOrgID(orgID))
	if err != nil {
		return fmt.Errorf("unable to create API client: %w", err)
	}

	installsCmds := &cobra.Command{
		Use:   "installs",
		Short: "View and manage installs of your app",
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long:    "List all your app's installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			installs, err := apiClient.GetAppInstalls(ctx, appID)
			if err != nil {
				return err
			}

			if len(installs) == 0 {
				ui.Line(ctx, "No installs of this app found. Create one using the nuon installs create command")
			} else {
				for _, install := range installs {
					statusColor := ui.GetStatusColor(install.Status)

					ui.Line(ctx, "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
				}
			}

			return nil
		},
	}
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			install, err := apiClient.GetInstall(ctx, id)
			if err != nil {
				return err
			}

			statusColor := ui.GetStatusColor(install.Status)

			ui.Line(ctx, "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
			return nil
		},
	}
	getCmd.PersistentFlags().StringVar(&id, "id", "", "Install ID")
	getCmd.MarkPersistentFlagRequired("id")
	installsCmds.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long:  "Create a new install of your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			install, err := apiClient.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
				Name: &name,
				AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
					Region:     region,
					IamRoleArn: &arn,
				},
			})
			if err != nil {
				return err
			}

			ui.Line(ctx, "Created new install: %s - %s", install.ID, install.Name)
			return nil
		},
	}
	createCmd.PersistentFlags().StringVar(&name, "name", "", "Install name")
	createCmd.MarkPersistentFlagRequired("name")
	createCmd.PersistentFlags().StringVar(&arn, "iam-arn", "", "IAM ARN")
	createCmd.MarkPersistentFlagRequired("iam-arn")
	createCmd.PersistentFlags().StringVar(&region, "aws-region", "", "AWS region")
	createCmd.MarkPersistentFlagRequired("aws-region")
	installsCmds.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := apiClient.DeleteInstall(ctx, id)
			if err != nil {
				return err
			}

			ui.Line(ctx, "Install %s was deleted", id)
			return nil
		},
	}
	deleteCmd.PersistentFlags().StringVar(&id, "id", "", "Install ID")
	deleteCmd.MarkPersistentFlagRequired("id")
	installsCmds.AddCommand(deleteCmd)

	rootCmd.AddCommand(installsCmds)

	return nil
}
