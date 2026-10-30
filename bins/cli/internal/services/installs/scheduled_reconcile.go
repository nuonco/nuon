package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// scheduledReconcileResult is the JSON payload for a scheduled-reconcile
// mutation (create or cancel).
type scheduledReconcileResult struct {
	InstallID   string `json:"install_id"`
	ID          string `json:"id"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	PlanOnly    bool   `json:"plan_only"`
	Status      string `json:"status"`
}

// CreateScheduledReconcile schedules a reconcile-app-version for the install
// at a future time instead of running it immediately.
func (s *Service) CreateScheduledReconcile(ctx context.Context, installID, at string, planOnly bool, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	scheduledAt, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return ui.PrintError(&ui.CLIUserError{
			Msg: fmt.Sprintf("--at must be an RFC3339 timestamp (for example 2026-01-02T15:04:05Z): %v", err),
		})
	}

	resp, err := s.api.CreateInstallScheduledReconcile(ctx, installID, scheduledAt.UTC().Format(time.RFC3339), planOnly)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	if asJSON {
		ui.PrintJSON(scheduledReconcileResult{
			InstallID:   installID,
			ID:          resp.ID,
			ScheduledAt: resp.ScheduledAt,
			PlanOnly:    planOnly,
			Status:      "scheduled",
		})
		return nil
	}

	ui.PrintLn(fmt.Sprintf(
		"scheduled reconcile %s for install %s at %s (local) / %s (UTC)",
		resp.ID, installID, scheduledAt.Local().Format(time.RFC3339), scheduledAt.UTC().Format(time.RFC3339),
	))
	ui.PrintLn(fmt.Sprintf("cancel with: nuon installs scheduled-reconciles cancel %s --install-id %s", resp.ID, installID))

	return nil
}

// ListScheduledReconciles lists the scheduled reconcile-app-version requests
// for an install, optionally filtered by status.
func (s *Service) ListScheduledReconciles(ctx context.Context, installID, status string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	schedules, err := s.api.GetInstallScheduledReconciles(ctx, installID, status, nil, nil)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(schedules)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"SCHEDULED AT",
			"STATUS",
			"PLAN ONLY",
			"REASON",
			"WORKFLOW ID",
		},
	}
	for _, sched := range schedules {
		data = append(data, []string{
			sched.ID,
			sched.ScheduledAt,
			string(sched.Status),
			fmt.Sprintf("%t", sched.PlanOnly),
			sched.StatusReason,
			sched.WorkflowID,
		})
	}
	view.Render(data)
	return nil
}

// CancelScheduledReconcile cancels a pending scheduled reconcile-app-version
// request for an install.
func (s *Service) CancelScheduledReconcile(ctx context.Context, installID, scheduledReconcileID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	resp, err := s.api.CancelInstallScheduledReconcile(ctx, installID, scheduledReconcileID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	status := "cancelled"
	var r models.AppInstallScheduledReconcile
	if resp != nil {
		r = *resp
		status = string(resp.Status)
	}

	printActionResult(asJSON, fmt.Sprintf("successfully cancelled scheduled reconcile %s", scheduledReconcileID), actionResult{
		InstallID: installID,
		ID:        r.ID,
		Status:    status,
	})

	return nil
}
