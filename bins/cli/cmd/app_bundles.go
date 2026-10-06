package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
)

func (c *cli) appBundlesCmd() *cobra.Command {
	var appID string
	group := &cobra.Command{Use: "bundles", Short: "Create and download immutable app bundles"}
	group.PersistentFlags().StringVarP(&appID, "app-id", "a", "", "App ID or name (default: selected app)")
	var query apps.BundleListOptions
	list := &cobra.Command{
		Use: "list", Short: "List app bundles, newest first", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.ListBundles(cmd.Context(), appID, query, PrintJSON)
		}),
	}
	list.Flags().StringVar(&query.ConfigID, "config-id", "", "Filter by exact app config ID")
	list.Flags().StringVar(&query.Status, "status", "", "Filter by queued, publishing, active, or error")
	list.Flags().IntVar(&query.Offset, "offset", 0, "Pagination offset")
	list.Flags().IntVar(&query.Limit, "limit", 20, "Page size (1–100)")
	group.AddCommand(list)
	var configID, platform string
	create := &cobra.Command{
		Use: "create", Short: "Start bundle publication and return immediately", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.CreateBundle(cmd.Context(), appID, configID, platform, PrintJSON)
		}),
	}
	create.Flags().StringVar(&configID, "config-id", "", "App config ID (default: latest successfully synced config)")
	create.Flags().StringVar(&platform, "platform", "linux/amd64", "Target platform")
	group.AddCommand(create)
	var getID string
	get := &cobra.Command{
		Use: "get", Short: "Inspect a bundle's status and archive metadata", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.GetBundle(cmd.Context(), appID, getID, PrintJSON)
		}),
	}
	get.Flags().StringVar(&getID, "bundle-id", "", "Bundle ID")
	_ = get.MarkFlagRequired("bundle-id")
	group.AddCommand(get)
	var waitID string
	var timeout time.Duration
	wait := &cobra.Command{
		Use: "wait", Short: "Wait for an existing bundle to finish publishing", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.WaitBundle(cmd.Context(), appID, waitID, timeout, PrintJSON)
		}),
	}
	wait.Flags().StringVar(&waitID, "bundle-id", "", "Bundle ID")
	wait.Flags().DurationVar(&timeout, "timeout", 30*time.Minute, "Maximum wait; publication continues after timeout or Ctrl+C")
	_ = wait.MarkFlagRequired("bundle-id")
	group.AddCommand(wait)
	var downloadID, file string
	var force bool
	download := &cobra.Command{
		Use: "download", Short: "Download and verify a published bundle; interrupted downloads resume automatically", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.apps.DownloadBundle(cmd.Context(), appID, downloadID, file, force, PrintJSON)
		}),
	}
	download.Flags().StringVar(&downloadID, "bundle-id", "", "Bundle ID")
	download.Flags().StringVar(&file, "file", "", "Destination archive path (default: server-provided filename)")
	download.Flags().BoolVar(&force, "force", false, "Replace an existing destination archive")
	_ = download.MarkFlagRequired("bundle-id")
	group.AddCommand(download)
	return group
}
