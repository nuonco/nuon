package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/builds"
	"github.com/spf13/cobra"
)

// newBuildsCmd constructs a new builds command
func newBuildsCmd(bindConfig bindConfigFunc, buildsService *builds.Service) *cobra.Command {
	var (
		buildID string
		compID  string
	)

	buildsCmd := &cobra.Command{
		Use:   "builds",
		Short: "Manage component builds",
		Long:  "Manage builds of your app components",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List builds",
		Long:    "List your app's builds",
		Run: func(cmd *cobra.Command, args []string) {
			buildsService.List(cmd.Context(), compID)
		},
	}
	listCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID of a component to filter builds by")
	listCmd.MarkFlagRequired("component-id")
	buildsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		Run: func(cmd *cobra.Command, args []string) {
			buildsService.Get(cmd.Context(), compID, buildID)
		},
	}
	getCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID of the component whose build you want to view")
	getCmd.MarkFlagRequired("component-id")
	getCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to view")
	getCmd.MarkFlagRequired("build-id")
	buildsCmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a build",
		Long:  "Create a build of an app component",
		Run: func(cmd *cobra.Command, args []string) {
			buildsService.Create(cmd.Context(), compID)
		},
	}
	createCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID of the component you want to create a build for")
	createCmd.MarkFlagRequired("component-id")
	buildsCmd.AddCommand(createCmd)

	return buildsCmd
}
