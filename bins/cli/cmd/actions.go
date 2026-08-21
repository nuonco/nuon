package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/actions"
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
			svc := c.actions
			return svc.List(cmd.Context(), appID, offset, limit, PrintJSON)
		}),
	}

	listCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter action workflows by")
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	actionsCmd.AddCommand(listCmd)

	installID := ""
	actionWorkflowID := ""
	roleName := ""
	recentRunsCmd := &cobra.Command{
		Use:   "recent-runs",
		Short: "Get action's most recent runs",
		Long:  "Get action's most recent runs for an install",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.actions
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
			svc := c.actions
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
			svc := c.actions
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
			svc := c.actions
			return svc.CreateRun(cmd.Context(), installID, actionWorkflowID, roleName, PrintJSON)
		}),
	}

	runCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID of the install you want to view recent runs for")
	runCmd.MarkFlagRequired("install-id")
	runCmd.Flags().StringVarP(&actionWorkflowID, "action-workflow-id", "w", "", "The ID of the action workflow you want to view recent runs for")
	runCmd.MarkFlagRequired("action-workflow-id")
	runCmd.Flags().StringVar(&roleName, "role-name", "", "IAM role name to use for action workflow")
	actionsCmd.AddCommand(runCmd)

	var adhocParams actions.AdHocParams
	adhocCmd := &cobra.Command{
		Use:         "adhoc",
		Short:       "Run a one-time action on an install",
		Args:        cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Example: `  nuon actions adhoc --command 'kubectl get pods'
  nuon actions adhoc --script ./debug.sh --env-file .env --env DEBUG=true
  nuon actions adhoc --command 'kubectl get pods' --wait`,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.actions.CreateAdHocRun(cmd.Context(), adhocParams, PrintJSON)
		}),
	}
	adhocCmd.Flags().StringVarP(&adhocParams.InstallID, "install-id", "i", "", "The ID or name of the install (defaults to the selected install)")
	adhocCmd.Flags().StringVar(&adhocParams.Command, "command", "", "Single-line shell command to execute")
	adhocCmd.Flags().StringVar(&adhocParams.ScriptPath, "script", "", "Path to a script to execute")
	adhocCmd.Flags().StringArrayVar(&adhocParams.Env, "env", nil, "Environment variable as KEY=VALUE (repeatable)")
	adhocCmd.Flags().StringVar(&adhocParams.EnvFile, "env-file", "", "Path to a dotenv-style environment file")
	adhocCmd.Flags().DurationVar(&adhocParams.Timeout, "timeout", 5*time.Minute, "Execution timeout between 1s and 1h")
	adhocCmd.Flags().StringVar(&adhocParams.Name, "name", "", "Display name for the action")
	adhocCmd.Flags().StringVar(&adhocParams.Role, "role", "", "IAM role to use for the action")
	adhocCmd.Flags().BoolVar(&adhocParams.EnableKubeConfig, "enable-kube-config", true, "Provide Kubernetes configuration to the action")
	adhocCmd.Flags().BoolVar(&adhocParams.Wait, "wait", false, "Wait for completion and print raw action logs")
	actionsCmd.AddCommand(adhocCmd)

	return actionsCmd
}
