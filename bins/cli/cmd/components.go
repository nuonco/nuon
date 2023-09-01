package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/components"
	"github.com/spf13/cobra"
)

func registerComponents(componentsService *components.Service) cobra.Command {
	var (
		buildID string
		id      string
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return componentsService.List(cmd.Context(), appID)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter components by")
	componentsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return componentsService.Get(cmd.Context(), id)
		},
	}
	getCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component you want to view")
	getCmd.MarkFlagRequired("component-id")
	componentsCmd.AddCommand(getCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete component",
		Long:  "Delete app component by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return componentsService.Delete(cmd.Context(), id)
		},
	}
	deleteCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component you want to delete")
	deleteCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(deleteCmd)

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build component",
		Long:  "Build a component by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return componentsService.Build(cmd.Context(), id)
		},
	}
	buildCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component you want to create a build for")
	buildCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(buildCmd)

	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Release a component build",
		Long:  "Release a component build by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return componentsService.Release(cmd.Context(), id, buildID)
		},
	}
	// TODO: Remove the componentID parameter from the SDK's CreateComponentRelease method,
	// so the user doesn't have to input the component id here.
	releaseCmd.Flags().StringVarP(&id, "component-id", "c", "", "The ID of the component who's build you want to release")
	releaseCmd.MarkFlagRequired("id")
	releaseCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to release")
	releaseCmd.MarkFlagRequired("id")
	componentsCmd.AddCommand(releaseCmd)

	return *componentsCmd
}
