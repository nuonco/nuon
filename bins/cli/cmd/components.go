package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/spf13/cobra"
)

func (c *cli) componentsCmd() *cobra.Command {
	var (
		id string
	)

	componentsCmd := &cobra.Command{
		Use:               "components",
		Short:             "Manage app components",
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
			svc.Get(cmd.Context(), id, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to view")
	getCmd.MarkFlagRequired("component-id")
	componentsCmd.AddCommand(getCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete component",
		Long:  "Delete app component by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.Delete(cmd.Context(), id, PrintJSON)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	deleteCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(deleteCmd)

	latestConfigCmd := &cobra.Command{
		Use:   "latest-config",
		Short: "Latest component config",
		Long:  "Show latest component config",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.LatestConfig(cmd.Context(), id, PrintJSON)
		},
	}
	latestConfigCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	latestConfigCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(latestConfigCmd)

	listConfigsCmd := &cobra.Command{
		Use:   "list-configs",
		Short: "List component configs",
		Long:  "List component configs",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := components.New(c.apiClient)
			svc.ListConfigs(cmd.Context(), id, PrintJSON)
		},
	}
	listConfigsCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	listConfigsCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(listConfigsCmd)

	return componentsCmd
}
