package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerReleases(ctx context.Context, api client.Client, cfg Config) cobra.Command {
	var (
		id        string
		installID string
	)

	releasesCmd := &cobra.Command{
		Use:   "releases",
		Short: "View and create releases of your app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List releases",
		Long:    "List releases of your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Line(ctx, "Not implemented")
			return nil
		},
	}
	releasesCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get release",
		Long:  "Get app release by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Line(ctx, "Not implemented")
			return nil
		},
	}
	getCmd.PersistentFlags().StringVar(&id, "id", "", "Release ID")
	getCmd.MarkPersistentFlagRequired("id")
	releasesCmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a release",
		Long:  "Create a new release of your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Line(ctx, "Not implemented")
			return nil
		},
	}
	releasesCmd.AddCommand(createCmd)

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "See release logs",
		Long:  "See release logs for an app install",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Line(ctx, "Not implemented")
			return nil
		},
	}
	logsCmd.PersistentFlags().StringVar(&id, "id", "", "Release ID")
	logsCmd.MarkPersistentFlagRequired("id")
	logsCmd.PersistentFlags().StringVar(&installID, "install-id", "", "Install ID")
	logsCmd.MarkPersistentFlagRequired("install-id")
	releasesCmd.AddCommand(logsCmd)

	return *releasesCmd
}
