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
		Short: "View your app's components",
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List components",
		Long:    "List your app's components",
		RunE: func(cmd *cobra.Command, args []string) error {
			components, err := c.api.GetAppComponents(ctx, c.cfg.APP_ID)
			if err != nil {
				return err
			}

			if len(components) == 0 {
				ui.Line(ctx, "No app components found. Create one using the nuon Terraform provider")
			} else {
				for _, component := range components {
					ui.Line(ctx, "%s - %s", component.ID, component.Name)
				}
			}

			return nil
		},
	}
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
	getCmd.PersistentFlags().StringVar(&id, "id", "", "Component ID")
	getCmd.MarkPersistentFlagRequired("id")
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
	deleteCmd.PersistentFlags().StringVar(&id, "id", "", "Component ID")
	deleteCmd.MarkPersistentFlagRequired("id")
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
	buildCmd.PersistentFlags().StringVar(&id, "id", "", "Component ID")
	buildCmd.MarkPersistentFlagRequired("id")
	componentsCmd.AddCommand(buildCmd)

	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Release a component build",
		Long:  "Release a component build by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			release, err := c.api.CreateComponentRelease(ctx, id, &models.ServiceCreateComponentReleaseRequest{
				BuildID: buildID,
				Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
					ReleaseStrategy: "parallel",
				},
			})
			if err != nil {
				return err
			}

			ui.Line(ctx, "Component release ID: %s", release.ID)
			return nil
		},
	}
	releaseCmd.PersistentFlags().StringVar(&id, "id", "", "Component ID")
	releaseCmd.MarkPersistentFlagRequired("id")
	releaseCmd.PersistentFlags().StringVar(&buildID, "build-id", "", "Build ID")
	releaseCmd.MarkPersistentFlagRequired("build-id")
	componentsCmd.AddCommand(releaseCmd)

	return *componentsCmd
}
