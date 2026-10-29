package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ReconcileAppVersion(ctx context.Context, installID string, planOnly bool, requestID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintLn("install id: " + installID)
	}

	resp, err := s.api.ReconcileInstallAppVersion(ctx, installID, planOnly, requestID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	workflowID := ""
	if resp != nil {
		workflowID = resp.WorkflowID
	}

	if s.cfg.Debug && workflowID != "" {
		ui.PrintLn("workflow id: " + workflowID)
	}

	status := "reconcile_app_version_triggered"
	if planOnly {
		status = "reconcile_app_version_planned"
	}

	id := ""
	if resp != nil {
		id = resp.ID
	}

	printActionResult(asJSON, "successfully triggered reconcile of install app version", actionResult{
		InstallID:  installID,
		ID:         id,
		WorkflowID: workflowID,
		Status:     status,
	})

	if !asJSON && workflowID != "" && s.cfg.Preview {
		return s.workflowsTUI(ctx, installID, workflowID, false)
	}

	return nil
}
