package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/runbooks"
)

func (c *cli) runbooksCmd() *cobra.Command {
	runbooksCmd := &cobra.Command{
		Use:               "runbooks",
		Short:             "Manage runbooks",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           InstallGroup.ID,
	}

	var (
		installID string
		runbookID string
		runID     string
		offset    int
		limit     int
	)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List runbooks",
		Long:    "List all runbooks for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), installID, PrintJSON)
		}),
	}
	listCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")
	listCmd.MarkFlagRequired("install-id")
	runbooksCmd.AddCommand(listCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a runbook",
		Long:  "Get a runbook by ID for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.Get(cmd.Context(), installID, runbookID, PrintJSON)
		}),
	}
	getCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")
	getCmd.MarkFlagRequired("install-id")
	getCmd.Flags().StringVarP(&runbookID, "runbook-id", "r", "", "The ID or name of the runbook")
	getCmd.MarkFlagRequired("runbook-id")
	runbooksCmd.AddCommand(getCmd)

	createRunCmd := &cobra.Command{
		Use:   "create-run",
		Short: "Run a runbook",
		Long:  "Trigger a runbook run by Install ID and Runbook ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.CreateRun(cmd.Context(), installID, runbookID, PrintJSON)
		}),
	}
	createRunCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")
	createRunCmd.MarkFlagRequired("install-id")
	createRunCmd.Flags().StringVarP(&runbookID, "runbook-id", "r", "", "The ID or name of the runbook")
	createRunCmd.MarkFlagRequired("runbook-id")
	runbooksCmd.AddCommand(createRunCmd)

	recentRunsCmd := &cobra.Command{
		Use:   "recent-runs",
		Short: "Get recent runbook runs",
		Long:  "List recent runbook runs for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.GetRecentRuns(cmd.Context(), installID, runbookID, offset, limit, PrintJSON)
		}),
	}
	recentRunsCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")
	recentRunsCmd.MarkFlagRequired("install-id")
	recentRunsCmd.Flags().StringVarP(&runbookID, "runbook-id", "r", "", "Filter runs by runbook ID or name")
	recentRunsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	recentRunsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	runbooksCmd.AddCommand(recentRunsCmd)

	getRunCmd := &cobra.Command{
		Use:   "get-run",
		Short: "Get a runbook run",
		Long:  "Get a runbook run by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.GetRun(cmd.Context(), installID, runID, PrintJSON)
		}),
	}
	getRunCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")
	getRunCmd.MarkFlagRequired("install-id")
	getRunCmd.Flags().StringVar(&runID, "run-id", "", "The ID of the run")
	getRunCmd.MarkFlagRequired("run-id")
	runbooksCmd.AddCommand(getRunCmd)

	return runbooksCmd
}
