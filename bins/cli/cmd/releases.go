package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/releases"
	"github.com/spf13/cobra"
)

func newReleasesCmd(bindConfig bindConfigFunc, releasesService *releases.Service) *cobra.Command {
	var (
		appID,
		compID,
		buildID,
		releaseID,
		delay string
		installsPerStep int64
		// installID string
	)

	releasesCmd := &cobra.Command{
		Use:   "releases",
		Short: "Manage releases",
		Long:  "Manages releases of app components to installs",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List releases",
		Long:    "List releases of a component",
		RunE: func(cmd *cobra.Command, args []string) error {
			return releasesService.List(cmd.Context(), appID, compID)
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter releases by")
	listCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID of a component to filter releases by")
	listCmd.MarkFlagsMutuallyExclusive("app-id", "component-id")
	releasesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get release",
		Long:  "Get an app release by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return releasesService.Get(cmd.Context(), releaseID)
		},
	}
	getCmd.Flags().StringVarP(&compID, "release-id", "r", "", "The ID of the release you want to view")
	getCmd.MarkFlagRequired("release-id")
	releasesCmd.AddCommand(getCmd)

	stepsCmd := &cobra.Command{
		Use:   "steps",
		Short: "Get release steps",
		Long:  "Get the steps for a release by release ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return releasesService.Steps(cmd.Context(), releaseID)
		},
	}
	stepsCmd.Flags().StringVarP(&releaseID, "release-id", "r", "", "The ID of the release whose steps you want to view")
	stepsCmd.MarkFlagRequired("release-id")
	releasesCmd.AddCommand(stepsCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create release",
		Long:  "Create a release of an app component",
		RunE: func(cmd *cobra.Command, args []string) error {
			return releasesService.Create(cmd.Context(), compID, buildID, delay, installsPerStep)
		},
	}
	createCmd.Flags().StringVarP(&compID, "component-id", "c", "", "The ID of the component whose build you want to create a release for")
	createCmd.MarkFlagRequired("component-id")
	createCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "The ID of the build you want to create a release for")
	createCmd.MarkFlagRequired("build-id")
	createCmd.Flags().StringVarP(&delay, "delay", "d", "10s", "The delay you want between each step")
	createCmd.Flags().Int64VarP(&installsPerStep, "installs-per-step", "i", 10, "The number of deploys you want to execute each step. Set to 0 to deploy to all installs in parrallel")
	releasesCmd.AddCommand(createCmd)

	// logsCmd := &cobra.Command{
	// 	Use:   "logs",
	// 	Short: "See release logs",
	// 	Long:  "See release logs for an app install",
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		ui.Line(cmd.Context(), "Not implemented")
	// 		return nil
	// 	},
	// }
	// logsCmd.PersistentFlags().StringVar(&releaseID, "id", "", "Release ID")
	// logsCmd.MarkPersistentFlagRequired("id")
	// logsCmd.PersistentFlags().StringVar(&installID, "install-id", "", "Install ID")
	// logsCmd.MarkPersistentFlagRequired("install-id")
	// releasesCmd.AddCommand(logsCmd)

	return releasesCmd
}
