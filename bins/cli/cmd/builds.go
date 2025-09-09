package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/builds"
)

// newBuildsCmd constructs a new builds command
func (c *cli) buildsCmd() *cobra.Command {
	var (
		buildID string
		compID  string
		appID   string
		offset  int
		limit   int
	)

	buildsCmd := &cobra.Command{
		Use:               "builds",
		Short:             "Manage component builds",
		Long:              "Manage builds of your app components",
		Aliases:           []string{"b"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List builds",
		Long:    "List your app's builds",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := builds.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), compID, appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of a component to filter builds by")
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	listCmd.MarkFlagRequired("app-id")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum builds to return")
	buildsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get build",
		Long:  "Get component build",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := builds.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), appID, compID, buildID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to view")
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	getCmd.MarkFlagRequired("app-id")
	getCmd.MarkFlagRequired("component-id")
	getCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to view")
	getCmd.MarkFlagRequired("build-id")
	buildsCmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a build",
		Long:  "Create a build of an app component",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := builds.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), appID, compID, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component you want to create a build for")
	createCmd.MarkFlagRequired("component-id")
	createCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	createCmd.MarkFlagRequired("app-id")
	buildsCmd.AddCommand(createCmd)

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View build logs",
		Long:  "View build logs by components and build ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := builds.New(c.apiClient, c.cfg)
			return svc.Logs(cmd.Context(), appID, compID, buildID, PrintJSON)
		}),
	}
	logsCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to view")
	logsCmd.MarkFlagRequired("component-id")
	logsCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The build ID for the build log you want to view")
	logsCmd.MarkFlagRequired("build-id")
	logsCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app the component belongs to")
	logsCmd.MarkFlagRequired("app-id")
	buildsCmd.AddCommand(logsCmd)

	return buildsCmd
}
