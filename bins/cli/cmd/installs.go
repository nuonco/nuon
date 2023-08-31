package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerInstalls(ctx context.Context) cobra.Command {
	var (
		id     string
		name   string
		arn    string
		region string
	)

	installsCmds := &cobra.Command{
		Use:   "installs",
		Short: "View and manage installs of your app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long:    "List all your app's installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			installs, err := c.api.GetAppInstalls(ctx, c.cfg.APP_ID)
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
			install, err := c.api.GetInstall(ctx, id)
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
			install, err := c.api.CreateInstall(ctx, c.cfg.APP_ID, &models.ServiceCreateInstallRequest{
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
			_, err := c.api.DeleteInstall(ctx, id)
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

	return *installsCmds
}
