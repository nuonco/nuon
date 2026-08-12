package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) SandboxRunGet(ctx context.Context, installID, runID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}
	view := ui.NewGetView()

	run, err := s.api.GetInstallSandboxRun(ctx, installID, runID)
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(run)
		return nil
	}

	rows := [][]string{
		{"install id", run.InstallID},
		{"run id", run.ID},
		{"run type", string(run.RunType)},
		{"status", run.StatusDescription},
		{"workflow id", run.WorkflowID},
	}
	view.Render(rows)
	return nil
}

func (s *Service) SandboxRunCancel(ctx context.Context, installID, runID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	run, err := s.api.GetInstallSandboxRun(ctx, installID, runID)
	if err != nil {
		return err
	}

	workflowID := run.WorkflowID
	if workflowID == "" {
		workflowID = run.InstallWorkflowID
	}
	if workflowID == "" {
		return ui.PrintError(fmt.Errorf("sandbox run %s has no associated workflow to cancel", runID))
	}

	if _, err := s.api.CancelWorkflow(ctx, workflowID); err != nil {
		return err
	}

	printActionResult(asJSON, fmt.Sprintf("successfully requested cancellation of sandbox run %s", runID), actionResult{
		InstallID:  installID,
		ID:         runID,
		WorkflowID: workflowID,
		Status:     "cancellation_requested",
	})
	return nil
}
