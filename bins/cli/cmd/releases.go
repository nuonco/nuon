package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/releases"
	"github.com/spf13/cobra"
)

func (c *cli) releasesCmd() *cobra.Command {
	var (
		appID,
		compID,
		buildID,
		releaseID,
		delay string
		installsPerStep int64
		autoBuild       bool
		latestBuild     bool
	)

	releasesCmd := &cobra.Command{
		Use:               "releases",
		Short:             "Manage releases",
		Long:              "Manages releases of app components to installs",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List releases",
		Long:    "List releases of a component",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := releases.New(c.apiClient)
			svc.List(cmd.Context(), appID, compID, PrintJSON)
		},
	}
	// TODO(ja): update cobra so we can require either app-id or component-id?
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter releases by")
	listCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of a component to filter releases by")
	listCmd.MarkFlagsOneRequired("app-id", "component-id")
	releasesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get release",
		Long:  "Get an app release by ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := releases.New(c.apiClient)
			svc.Get(cmd.Context(), releaseID, PrintJSON)
		},
	}
	getCmd.Flags().StringVarP(&releaseID, "release-id", "r", "", "The ID of the release you want to view")
	getCmd.MarkFlagRequired("release-id")
	releasesCmd.AddCommand(getCmd)

	stepsCmd := &cobra.Command{
		Use:   "steps",
		Short: "Get release steps",
		Long:  "Get the steps for a release by release ID",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := releases.New(c.apiClient)
			svc.Steps(cmd.Context(), releaseID, PrintJSON)
		},
	}
	stepsCmd.Flags().StringVarP(&releaseID, "release-id", "r", "", "The ID of the release whose steps you want to view")
	stepsCmd.MarkFlagRequired("release-id")
	releasesCmd.AddCommand(stepsCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create release",
		Long:  "Create a release of an app component",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := releases.New(c.apiClient)
			svc.Create(cmd.Context(), compID, buildID, delay, autoBuild, latestBuild, installsPerStep, PrintJSON)
		},
	}
	createCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID or name of the component whose build you want to create a release for")
	createCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to create a release for")
	createCmd.MarkFlagsOneRequired("build-id", "component-id")
	createCmd.Flags().StringVarP(&delay, "delay", "d", "10s", "The delay you want between each step")
	createCmd.Flags().Int64VarP(&installsPerStep, "installs-per-step", "s", 10, "The number of deploys you want to execute each step. Set to 0 to deploy to all installs in parrallel")
	createCmd.Flags().BoolVarP(&autoBuild, "auto-build", "", false, "automatically trigger a build for the provided component")
	createCmd.Flags().BoolVarP(&latestBuild, "latest-build", "", false, "uses the latest build for a component")
	releasesCmd.AddCommand(createCmd)

	// logsCmd := &cobra.Command{
	//	Use:   "logs",
	//	Short: "See release logs",
	//	Long:  "See release logs for an app install",
	//	Run: func(cmd *cobra.Command, args []string) {
	//		ui.Line(cmd.Context(), "Not implemented")
	//		nil
	//	},
	// }
	// logsCmd.PersistentFlags().StringVar(&releaseID, "id", "", "Release ID")
	// logsCmd.MarkPersistentFlagRequired("id")
	// logsCmd.PersistentFlags().StringVar(&installID, "install-id", "", "Install ID")
	// logsCmd.MarkPersistentFlagRequired("install-id")
	// releasesCmd.AddCommand(logsCmd)

	return releasesCmd
}
