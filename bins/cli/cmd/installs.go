package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/spf13/cobra"
)

func registerInstalls(installsService *installs.Service) cobra.Command {
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
			return installsService.List(cmd.Context(), appID)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter installs by")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installsService.Get(cmd.Context(), id)
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
			return installsService.Create(cmd.Context(), appID, name, region, arn)
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
			return installsService.Delete(cmd.Context(), id)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID of the install you want to view")
	deleteCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deleteCmd)

	return *installsCmds
}
