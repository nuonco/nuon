package actions

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type CreateRunParams struct {
	InstallID        string
	ActionWorkflowID string
	Role             string
	RequestID        string
	Env              []string
	EnvFile          string
	Wait             bool
}

type createRunResult struct {
	WorkflowID       string `json:"workflow_id"`
	InstallID        string `json:"install_id"`
	ActionWorkflowID string `json:"action_workflow_id"`
}

func (s *Service) CreateRun(ctx context.Context, params CreateRunParams, asJSON bool) error {
	view := ui.NewGetView()

	install, err := s.installOrSelected(ctx, params.InstallID)
	if err != nil {
		return view.Error(err)
	}
	installID := install.ID

	actionWorkflowID, err := lookup.ActionWorkflowID(ctx, s.api, install.AppID, params.ActionWorkflowID)
	if err != nil {
		return view.Error(err)
	}

	if params.Role != "" {
		roles, err := s.api.GetAvailableRoles(ctx, installID)
		if err != nil {
			return view.Error(fmt.Errorf("get available roles: %w", err))
		}
		if err := validateRole(params.Role, roles); err != nil {
			return view.Error(err)
		}
	}

	awc, err := s.api.GetActionWorkflowLatestConfig(ctx, actionWorkflowID)
	if err != nil {
		return view.Error(fmt.Errorf("error getting action workflow config: %w", err))
	}

	envVars, err := parseRunEnv(params.EnvFile, params.Env)
	if err != nil {
		return view.Error(err)
	}

	req := &models.ServiceCreateInstallActionWorkflowRunRequest{
		ActionWorkflowConfigID: &awc.ID,
		Role:                   params.Role,
		RequestID:              params.RequestID,
	}
	if len(envVars) > 0 {
		req.RunEnvVars = envVars
	}

	resp, err := s.api.CreateInstallActionWorkflowRunWithResponse(ctx, installID, req)
	if err != nil {
		return view.Error(fmt.Errorf("error creating action workflow run: %w", err))
	}
	result := createRunResult{
		WorkflowID:       resp.WorkflowID,
		InstallID:        installID,
		ActionWorkflowID: actionWorkflowID,
	}

	if params.Wait {
		logWriter := io.Writer(os.Stdout)
		if asJSON {
			logWriter = os.Stderr
		}

		spinner := bubbles.NewSpinnerView(asJSON, s.cfg.Interactive)
		spinner.Start("waiting for action run to be created")
		runID, err := s.waitForWorkflowRun(ctx, installID, actionWorkflowID, resp.WorkflowID, spinner)
		if err != nil {
			spinner.Stop()
			return printWaitError(err, asJSON)
		}
		completedRun, err := s.waitForRun(ctx, installID, runID, logWriter, spinner)
		spinner.Stop()
		if err != nil {
			return printWaitError(err, asJSON)
		}
		if completedRun.Status != string(models.AppInstallActionWorkflowRunStatusFinished) {
			message := fmt.Sprintf("action run %s ended with status %s", completedRun.ID, completedRun.Status)
			if completedRun.StatusDescription != "" {
				message += ": " + completedRun.StatusDescription
			}
			return printWaitError(&ui.CLIUserError{Msg: message}, asJSON)
		}
		if asJSON {
			ui.PrintJSON(completedRun)
			return nil
		}
		spinner.Success(fmt.Sprintf("action run %s finished", completedRun.ID))
		renderCompletedRun(completedRun)
		return nil
	}

	if asJSON {
		ui.PrintJSON(result)
		return nil
	}

	ui.PrintSuccess("action run queued")
	view.Render([][]string{
		{"workflow id", result.WorkflowID},
		{"install id", result.InstallID},
		{"action workflow id", result.ActionWorkflowID},
	})

	return nil
}

// waitForWorkflowRun polls until the run created by the install workflow
// appears, since the run row is written asynchronously.
func (s *Service) waitForWorkflowRun(ctx context.Context, installID, actionWorkflowID, workflowID string, spinner *bubbles.SpinnerView) (string, error) {
	for {
		runID, err := s.findWorkflowRun(ctx, installID, actionWorkflowID, workflowID)
		if err != nil || runID != "" {
			return runID, err
		}

		workflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return "", fmt.Errorf("get workflow: %w", err)
		}
		if workflow.Status != nil && isTerminalWorkflowStatus(workflow.Status.Status) {
			// The run may have been written between the runs check and the workflow check.
			runID, err := s.findWorkflowRun(ctx, installID, actionWorkflowID, workflowID)
			if err != nil || runID != "" {
				return runID, err
			}
			return "", &ui.CLIUserError{
				Msg: fmt.Sprintf("action workflow %s ended with status %s before a run was created", workflowID, workflow.Status.Status),
			}
		}
		if workflow.Status != nil && workflow.Status.Status != "" {
			spinner.Update(fmt.Sprintf("waiting for action run to be created (workflow status: %s)", workflow.Status.Status))
		}

		if err := waitForRunPoll(ctx); err != nil {
			return "", err
		}
	}
}

func (s *Service) findWorkflowRun(ctx context.Context, installID, actionWorkflowID, workflowID string) (string, error) {
	iaw, _, err := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflowID, &models.GetPaginatedQuery{Limit: 10})
	if err != nil {
		return "", fmt.Errorf("get action runs: %w", err)
	}
	if iaw == nil {
		return "", nil
	}
	for _, run := range iaw.Runs {
		if run.InstallWorkflowID == workflowID || run.WorkflowID == workflowID {
			return run.ID, nil
		}
	}
	return "", nil
}

func isTerminalWorkflowStatus(status models.AppStatus) bool {
	switch status {
	case models.AppStatusSuccess,
		models.AppStatusError,
		models.AppStatusCancelled,
		models.AppStatusDiscarded,
		models.AppStatusNotDashAttempted,
		models.AppStatusUserDashSkipped,
		models.AppStatusAutoDashSkipped,
		models.AppStatusApprovalDashExpired,
		models.AppStatusApprovalDashDenied,
		models.AppStatusNoop:
		return true
	default:
		return false
	}
}
