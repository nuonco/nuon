package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/components"
)

func (c *cli) componentsCmd() *cobra.Command {
	var id string

	componentsCmd := &cobra.Command{
		Use:               "components",
		Short:             "Manage app components",
		Aliases:           []string{"c"},
		PersistentPreRunE: c.persistentPreRunE,
	}

	appID := ""
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List components",
		Long:    "List your app's components",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.List(cmd.Context(), appID, PrintJSON)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter components by")
	componentsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.Get(cmd.Context(), appID, id, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to view")
	getCmd.MarkFlagRequired("component-id")
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	getCmd.MarkFlagRequired("app-id")
	componentsCmd.AddCommand(getCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete component",
		Long:  "Delete app component by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.Delete(cmd.Context(), appID, id, PrintJSON)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	deleteCmd.MarkFlagRequired("id")
	deleteCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	deleteCmd.MarkFlagRequired("app-id")
	componentsCmd.AddCommand(deleteCmd)

	latestConfigCmd := &cobra.Command{
		Use:   "latest-config",
		Short: "Latest component config",
		Long:  "Show latest component config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.LatestConfig(cmd.Context(), appID, id, PrintJSON)
		},
	}
	latestConfigCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	latestConfigCmd.MarkFlagRequired("id")
	latestConfigCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	latestConfigCmd.MarkFlagRequired("app-id")
	componentsCmd.AddCommand(latestConfigCmd)

	listConfigsCmd := &cobra.Command{
		Use:   "list-configs",
		Short: "List component configs",
		Long:  "List component configs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.ListConfigs(cmd.Context(), appID, id, PrintJSON)
		},
	}
	listConfigsCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	listConfigsCmd.MarkFlagRequired("id")
	listConfigsCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	listConfigsCmd.MarkFlagRequired("app-id")
	componentsCmd.AddCommand(listConfigsCmd)

	return componentsCmd
}
