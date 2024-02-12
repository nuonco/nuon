package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/builds"
	"github.com/spf13/cobra"
)

// newBuildsCmd constructs a new builds command
func (c *cli) buildsCmd() *cobra.Command {
	var (
		buildID string
		compID  string
		appID   string
		limit   int64
	)

	buildsCmd := &cobra.Command{
		Use:               "builds",
		Short:             "Manage component builds",
		Long:              "Manage builds of your app components",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List builds",
		Long:    "List your app's builds",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := builds.New(c.apiClient)
			svc.List(cmd.Context(), compID, appID, &limit, PrintJSON)
		},
	}
	listCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of a component to filter builds by")
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the component to filter builds by")
	listCmd.Flags().Int64VarP(&limit, "limit", "l", 60, "Maximum health checks to return")
	buildsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get build",
		Long:  "Get component build",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := builds.New(c.apiClient)
			svc.Get(cmd.Context(), compID, buildID, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to view")
	getCmd.MarkFlagRequired("component-id")
	getCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to view")
	getCmd.MarkFlagRequired("build-id")
	buildsCmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a build",
		Long:  "Create a build of an app component",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := builds.New(c.apiClient)
			svc.Create(cmd.Context(), compID, PrintJSON)
		},
	}
	createCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component you want to create a build for")
	createCmd.MarkFlagRequired("component-id")
	buildsCmd.AddCommand(createCmd)

	printPlanCmd := &cobra.Command{
		Use:   "print-plan",
		Short: "Print build plan",
		Long:  "Print build plan",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := builds.New(c.apiClient)
			svc.PrintPlan(cmd.Context(), compID, buildID, PrintJSON)
		},
	}
	printPlanCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to view")
	printPlanCmd.MarkFlagRequired("component-id")
	printPlanCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The build ID for the build plan you want to print")
	printPlanCmd.MarkFlagRequired("build-id")
	buildsCmd.AddCommand(printPlanCmd)

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View build logs",
		Long:  "View build logs by components and build ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := builds.New(c.apiClient)
			svc.Logs(cmd.Context(), compID, buildID, PrintJSON)
		},
	}
	logsCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to view")
	logsCmd.MarkFlagRequired("component-id")
	logsCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The build ID for the build log you want to view")
	logsCmd.MarkFlagRequired("build-id")
	buildsCmd.AddCommand(logsCmd)

	return buildsCmd
}
