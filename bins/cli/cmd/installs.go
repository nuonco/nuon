package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/spf13/cobra"
)

func (c *cli) installsCmd() *cobra.Command {
	var (
		id     string
		name   string
		arn    string
		region string
		appID  string
	)

	installsCmds := &cobra.Command{
		Use:               "installs",
		Short:             "Manage app installs",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installs",
		Long:    "List all your app's installs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient)
			svc.List(cmd.Context(), appID, PrintJSON)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	installsCmds.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an install",
		Long:  "Get an install by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient)
			svc.Get(cmd.Context(), id, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	getCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an install",
		Long:  "Create a new install of your app",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient)
			svc.Create(cmd.Context(), appID, name, region, arn, PrintJSON)
		},
	}
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app to create this install for")
	createCmd.MarkFlagRequired("app-id")
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name you want to give this install")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&arn, "role", "o", "", "The ARN of the IAM role to use to provision this install")
	createCmd.MarkFlagRequired("iam-arn")
	createCmd.Flags().StringVarP(&region, "region", "r", "", "The region to provision this install in")
	createCmd.MarkFlagRequired("region")
	installsCmds.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete install",
		Long:  "Delete an install by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient)
			svc.Delete(cmd.Context(), id, PrintJSON)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	deleteCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(deleteCmd)

	componentsCmd := &cobra.Command{
		Use:   "components",
		Short: "Get install components",
		Long:  "Get all components on an install",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installs.New(c.apiClient)
			svc.Components(cmd.Context(), id, PrintJSON)
		},
	}
	componentsCmd.Flags().StringVarP(&id, "install-id", "i", "", "The ID or name of the install you want to view")
	componentsCmd.MarkFlagRequired("install-id")
	installsCmds.AddCommand(componentsCmd)

	return installsCmds
}
