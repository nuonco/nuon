package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateInstallRunbookRunWorkflowInput struct {
	InstallID       string            `json:"install_id"`
	RunbookID       string            `json:"runbook_id"`
	RunbookConfigID string            `json:"runbook_config_id"`
	TriggeredByID   string            `json:"triggered_by_id"`
	Inputs          map[string]string `json:"inputs,omitempty"`
	Callback        callback.Ref      `json:"callback,omitempty"`

	// IdempotencyKey must be deterministic for a given (branch run, install,
	// runbook position). Without it a Temporal retry of this activity after the
	// underlying transaction has committed starts the runbook a second time.
	IdempotencyKey string `json:"idempotency_key"`
}

type CreateInstallRunbookRunWorkflowOutput struct {
	WorkflowID          string `json:"workflow_id"`
	InstallRunbookRunID string `json:"install_runbook_run_id"`

	// TerminalStatus is set when the run had already finished, so no completion
	// signal is coming and the caller must not wait on the callback. It is not
	// necessarily success.
	TerminalStatus string `json:"terminal_status,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallRunbookRunWorkflow(ctx context.Context, input *CreateInstallRunbookRunWorkflowInput) (*CreateInstallRunbookRunWorkflowOutput, error) {
	if input.RunbookID == "" || input.RunbookConfigID == "" {
		return nil, fmt.Errorf("runbook id and runbook config id are required to create a runbook run")
	}
	if input.IdempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required so an activity retry does not run the runbook twice")
	}

	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var installRunbook app.InstallRunbook
	if err := a.db.WithContext(ctx).
		Where(app.InstallRunbook{InstallID: input.InstallID, RunbookID: input.RunbookID}).
		First(&installRunbook).Error; err != nil {
		return nil, fmt.Errorf("unable to get install runbook: %w", err)
	}

	// Activities carry no request auth context, so set the created-by/org the
	// InstallRunbookRun BeforeCreate hooks read.
	ctx = cctx.SetAccountIDContext(ctx, input.TriggeredByID)
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)

	triggered, err := a.runbooksHelpers.TriggerRunbookRun(ctx, runbookshelpers.TriggerRunbookRunRequest{
		InstallRunbookID: installRunbook.ID,
		RunbookConfigID:  input.RunbookConfigID,
		TriggeredByID:    input.TriggeredByID,
		Inputs:           input.Inputs,
		Callback:         input.Callback,
		IdempotencyKey:   &input.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to trigger runbook run: %w", err)
	}

	return &CreateInstallRunbookRunWorkflowOutput{
		WorkflowID:          triggered.Workflow.ID,
		InstallRunbookRunID: triggered.Run.ID,
		TerminalStatus:      triggered.TerminalStatus,
	}, nil
}
