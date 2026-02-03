package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/policies"
)

func (c *cli) policiesCmd() *cobra.Command {
	var (
		appID     string
		configID  string
		reportID  string
		installID string
		ownerType string
		status    string
		format    string
		output    string
		offset    int
		limit     int
	)

	policiesCmd := &cobra.Command{
		Use:               "policies",
		Short:             "Manage app policies and reports",
		Long:              "Manage app policies configurations and view policy reports",
		Aliases:           []string{"pol"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	// List policies configs
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List policies configs",
		Long:    "List your app's policies configurations",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := policies.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app")
	listCmd.MarkFlagRequired("app-id")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum results to return")
	policiesCmd.AddCommand(listCmd)

	// Get policies config
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get policies config",
		Long:  "Get a specific or latest policies configuration",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := policies.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), appID, configID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of the app")
	getCmd.MarkFlagRequired("app-id")
	getCmd.Flags().StringVarP(&configID, "config-id", "c", "", "The ID of the policies config (omit for latest)")
	policiesCmd.AddCommand(getCmd)

	// Reports subcommand
	reportsCmd := &cobra.Command{
		Use:     "reports",
		Aliases: []string{"rep"},
		Short:   "Manage policy reports",
		Long:    "View and export policy evaluation reports",
	}
	policiesCmd.AddCommand(reportsCmd)

	// List reports
	reportsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List policy reports",
		Long:    "List policy reports with optional filters",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := policies.New(c.apiClient, c.cfg)
			return svc.ListReports(cmd.Context(), appID, installID, ownerType, status, offset, limit, PrintJSON)
		}),
	}
	reportsListCmd.Flags().StringVarP(&appID, "app-id", "a", "", "Filter by app ID or name")
	reportsListCmd.Flags().StringVarP(&installID, "install-id", "i", "", "Filter by install ID or name")
	reportsListCmd.Flags().StringVar(&ownerType, "owner-type", "", "Filter by owner type (install_deploys, install_sandbox_runs, component_builds)")
	reportsListCmd.Flags().StringVar(&status, "status", "", "Filter by status")
	reportsListCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	reportsListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum results to return")
	reportsCmd.AddCommand(reportsListCmd)

	// Get report
	reportsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get policy report",
		Long:  "Get details of a specific policy report",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := policies.New(c.apiClient, c.cfg)
			return svc.GetReport(cmd.Context(), reportID, PrintJSON)
		}),
	}
	reportsGetCmd.Flags().StringVarP(&reportID, "report-id", "r", "", "The ID of the policy report")
	reportsGetCmd.MarkFlagRequired("report-id")
	reportsCmd.AddCommand(reportsGetCmd)

	// Export report
	reportsExportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export policy report",
		Long:  "Export a policy report in OPA, SARIF, or PDF format",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := policies.New(c.apiClient, c.cfg)
			return svc.ExportReport(cmd.Context(), reportID, format, output, PrintJSON)
		}),
	}
	reportsExportCmd.Flags().StringVarP(&reportID, "report-id", "r", "", "The ID of the policy report")
	reportsExportCmd.MarkFlagRequired("report-id")
	reportsExportCmd.Flags().StringVarP(&format, "format", "F", "opa", "Export format: opa, sarif, or pdf")
	reportsExportCmd.Flags().StringVarP(&output, "output", "O", "", "Output file path (required for PDF)")
	reportsCmd.AddCommand(reportsExportCmd)

	return policiesCmd
}
