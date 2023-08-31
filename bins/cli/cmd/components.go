package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerComponents(ctx context.Context) cobra.Command {
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
			components := []*models.AppComponent{}
			err := error(nil)
			if appID != "" {
				components, err = c.api.GetAppComponents(ctx, appID)
			} else {
				components, err = c.api.GetAllComponents(ctx)
			}
			if err != nil {
				return err
			}

			if len(components) == 0 {
				ui.Line(ctx, "No components found")
			} else {
				for _, component := range components {
					ui.Line(ctx, "%s - %s", component.ID, component.Name)
				}
			}

			return nil
		},
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID of an app to filter components by")
	componentsCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get component",
		Long:  "Get app component by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			component, err := c.api.GetComponent(ctx, id)
			if err != nil {
				return err
			}

			ui.Line(ctx, "%s - %s", component.ID, component.Name)
			return nil
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
			_, err := c.api.DeleteComponent(ctx, id)
			if err != nil {
				return err
			}

			ui.Line(ctx, "Component %s was deleted", id)
			return nil
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
			build, err := c.api.CreateComponentBuild(
				ctx, id, &models.ServiceCreateComponentBuildRequest{UseLatest: true})
			if err != nil {
				return err
			}

			ui.Line(ctx, "Component build ID: %s", build.ID)
			return nil
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
			release, err := c.api.CreateComponentRelease(ctx, id, &models.ServiceCreateComponentReleaseRequest{
				BuildID: buildID,
				Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
					InstallsPerStep: 0,
				},
			})
			if err != nil {
				return err
			}

			ui.Line(ctx, "Component release ID: %s", release.ID)
			return nil
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
