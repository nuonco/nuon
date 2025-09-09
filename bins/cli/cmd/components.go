package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/components"
)

func (c *cli) componentsCmd() *cobra.Command {
	var (
		id     string
		appID  string
		offset int
		limit  int
	)
	componentsCmd := &cobra.Command{
		Use:               "components",
		Short:             "Manage app components",
		Aliases:           []string{"c"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List components",
		Long:    "List your app's components",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := components.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter components by")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	componentsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := components.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), appID, id, PrintJSON)
		}),
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
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := components.New(c.apiClient, c.cfg)
			return svc.Delete(cmd.Context(), appID, id, PrintJSON)
		}),
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
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := components.New(c.apiClient, c.cfg)
			return svc.LatestConfig(cmd.Context(), appID, id, PrintJSON)
		}),
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
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := components.New(c.apiClient, c.cfg)
			return svc.ListConfigs(cmd.Context(), appID, id, offset, limit, PrintJSON)
		}),
	}
	listConfigsCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID or name of the component you want to delete")
	listConfigsCmd.MarkFlagRequired("id")
	listConfigsCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	listConfigsCmd.MarkFlagRequired("app-id")
	listConfigsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listConfigsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	componentsCmd.AddCommand(listConfigsCmd)

	return componentsCmd
}
