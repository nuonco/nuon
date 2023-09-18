package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/spf13/cobra"
)

func newComponentsCmd(bindConfig config.BindCobraFunc, componentsService *components.Service) *cobra.Command {
	var (
		id string
	)

	componentsCmd := &cobra.Command{
		Use:   "components",
		Short: "Manage app components",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	appID := ""
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List components",
		Long:    "List your app's components",
		Run: func(cmd *cobra.Command, args []string) {
			componentsService.List(cmd.Context(), appID, PrintJSON)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter components by")
	componentsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		Run: func(cmd *cobra.Command, args []string) {
			componentsService.Get(cmd.Context(), id, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component you want to view")
	getCmd.MarkFlagRequired("component-id")
	componentsCmd.AddCommand(getCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete component",
		Long:  "Delete app component by ID",
		Run: func(cmd *cobra.Command, args []string) {
			componentsService.Delete(cmd.Context(), id, PrintJSON)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component you want to delete")
	deleteCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(deleteCmd)

	return componentsCmd
}
