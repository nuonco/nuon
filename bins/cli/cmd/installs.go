package cmd

import (
	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerInstalls() cobra.Command {
	var (
		id     string
		name   string
		arn    string
		region string
		appID  string = ""
	)

	installsCmds := &cobra.Command{
		Use:   "installs",
		Short: "Manage app installs",
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
			installs := []*models.AppInstall{}
			err := error(nil)
			if appID == "" {
				installs, err = c.api.GetAllInstalls(cmd.Context())
			} else {
				installs, err = c.api.GetAppInstalls(cmd.Context(), appID)
			}
			if err != nil {
				return err
			}

			if len(installs) == 0 {
				ui.Line(cmd.Context(), "No installs of this app found")
			} else {
				for _, install := range installs {
					statusColor := ui.GetStatusColor(install.Status)
					ui.Line(cmd.Context(), "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
				}
			}

			return nil
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter installs by")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			install, err := c.api.GetInstall(cmd.Context(), id)
			if err != nil {
				return err
			}

			statusColor := ui.GetStatusColor(install.Status)
			ui.Line(cmd.Context(), "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
			return nil
		},
	}
	getCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to view")
	getCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long:  "Create a new install of your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			install, err := c.api.CreateInstall(cmd.Context(), appID, &models.ServiceCreateInstallRequest{
				Name: &name,
				AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
					Region:     region,
					IamRoleArn: &arn,
				},
			})
			if err != nil {
				return err
			}

			ui.Line(cmd.Context(), "Created new install: %s - %s", install.ID, install.Name)
			return nil
		},
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of the app to create this install for")
	createCmd.MarkFlagRequired("app-id")
	createCmd.Flags().StringVarP(&name, "install-name", "n", "", "The name you want to give this install")
	createCmd.MarkFlagRequired("install-name")
	createCmd.Flags().StringVarP(&arn, "install-iam-arn", "i", "", "The ARN of the role to use to provision this install")
	createCmd.MarkFlagRequired("install-iam-arn")
	createCmd.Flags().StringVarP(&region, "install-region", "r", "", "The region to provision this install in")
	createCmd.MarkFlagRequired("install-region")
	installsCmds.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := c.api.DeleteInstall(cmd.Context(), id)
			if err != nil {
				return err
			}

			ui.Line(cmd.Context(), "Install %s was deleted", id)
			return nil
		},
	}
	deleteCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to view")
	deleteCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deleteCmd)

	return *installsCmds
}
