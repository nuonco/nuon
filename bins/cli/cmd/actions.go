package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/actions"
)

func (c *cli) actionsCmd() *cobra.Command {
	var (
		offset int
		limit  int
	)

	actionsCmd := &cobra.Command{
		Use:               "actions",
		Short:             "Manage app actions",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	appID := ""
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all app actions",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := actions.New(c.v, c.apiClient, c.cfg)
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}

	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter action workflows by")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	actionsCmd.AddCommand(listCmd)

	installID := ""
	actionWorkflowID := ""
	recentRunsCmd := &cobra.Command{
		Use:   "recent-runs",
		Short: "Get action's most recent runs",
		Long:  "Get action's most recent runs for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := actions.New(c.v, c.apiClient, c.cfg)
			return svc.GetRecentRuns(cmd.Context(), installID, actionWorkflowID, offset, limit, PrintJSON)
		}),
	}
	recentRunsCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID of the install you want to view recent runs for")
	recentRunsCmd.MarkFlagRequired("install-id")
	recentRunsCmd.Flags().StringVarP(&actionWorkflowID, "action-workflow-id", "w", "", "The ID of the action workflow you want to view recent runs for")
	recentRunsCmd.MarkFlagRequired("action-workflow-id")
	recentRunsCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	recentRunsCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	actionsCmd.AddCommand(recentRunsCmd)

	deleteWorkflowCmd := &cobra.Command{
		Use:   "delete-workflow",
		Short: "Delete an action workflow",
		Long:  "Delete an action workflow by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := actions.New(c.v, c.apiClient, c.cfg)
			return svc.DeleteWorkflow(cmd.Context(), actionWorkflowID)
		}),
	}
	deleteWorkflowCmd.Flags().StringVarP(&actionWorkflowID, "action-workflow-id", "w", "", "The ID of the action workflow you want to delete")
	deleteWorkflowCmd.MarkFlagRequired("action-workflow-id")
	actionsCmd.AddCommand(deleteWorkflowCmd)

	runID := ""
	getRunCmd := &cobra.Command{
		Use:   "get-run",
		Short: "Get an action run",
		Long:  "Get an action run by ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := actions.New(c.v, c.apiClient, c.cfg)
			return svc.GetRun(cmd.Context(), installID, runID, PrintJSON)
		}),
	}
	getRunCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID of the install you want to view recent runs for")
	getRunCmd.MarkFlagRequired("install-id")
	getRunCmd.Flags().StringVarP(&runID, "run-id", "r", "", "The ID of the run you want to view")
	getRunCmd.MarkFlagRequired("run-id")
	actionsCmd.AddCommand(getRunCmd)

	runCmd := &cobra.Command{
		Use:   "create-run",
		Short: "Run an action",
		Long:  "Run an action by Install ID and Action Workflow ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := actions.New(c.v, c.apiClient, c.cfg)
			return svc.CreateRun(cmd.Context(), installID, actionWorkflowID, PrintJSON)
		}),
	}

	runCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID of the install you want to view recent runs for")
	runCmd.MarkFlagRequired("install-id")
	runCmd.Flags().StringVarP(&actionWorkflowID, "action-workflow-id", "w", "", "The ID of the action workflow you want to view recent runs for")
	runCmd.MarkFlagRequired("action-workflow-id")
	actionsCmd.AddCommand(runCmd)

	return actionsCmd
}
