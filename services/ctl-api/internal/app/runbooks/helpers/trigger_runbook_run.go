package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerRunbookRunRequest struct {
	InstallRunbookID, RunbookConfigID, TriggeredByID string
	Inputs                                           map[string]string
	StepSelections                                   []app.RunbookStepSelection
	TriggerEventDispatchID                           *string
	Role                                             string

	// IdempotencyKey makes the call safe to repeat: a second call with the same
	// key returns the existing run instead of starting the runbook again. Callers
	// driven by a retryable Temporal activity must set it, otherwise a retry after
	// a committed transaction runs the runbook twice.
	IdempotencyKey *string

	// Callback, when set, receives a Temporal signal once the run completes.
	//
	// enqueueSignal does NOT merge callbacks onto a deduplicated row — its
	// conflict clause is DoNothing, so a callback handed to an enqueue that
	// dedupes is discarded. The dedupe path below therefore appends the callback
	// to the existing signal explicitly, and reports TerminalStatus when the run
	// has already finished and no callback can fire. Callers that set Callback
	// must handle TerminalStatus, or they will block until FallbackAwaitTimeout.
	Callback callback.Ref
}

type TriggerRunbookRunResponse struct {
	Run           *app.InstallRunbookRun
	Workflow      *app.Workflow
	QueueSignalID string

	// TerminalStatus is set when a Callback was requested but the run had already
	// reached a terminal state, so no completion signal will be sent. It carries
	// that state, which is not necessarily success — a deduped run may already
	// have failed or been cancelled.
	TerminalStatus string
}

func (h *Helpers) TriggerRunbookRun(ctx context.Context, req TriggerRunbookRunRequest) (*TriggerRunbookRunResponse, error) {
	var run app.InstallRunbookRun
	var workflow app.Workflow
	var queueSignalID string
	var queueID string
	var terminalStatus string
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var installRunbook app.InstallRunbook
		if err := tx.Preload("Runbook").Preload("Install").Where(app.InstallRunbook{ID: req.InstallRunbookID}).First(&installRunbook).Error; err != nil {
			return err
		}
		var runbookConfig app.RunbookConfig
		if err := tx.Preload("Inputs").Where(app.RunbookConfig{
			ID: req.RunbookConfigID, OrgID: installRunbook.OrgID, RunbookID: installRunbook.RunbookID,
		}).First(&runbookConfig).Error; err != nil {
			return err
		}
		supplied := make(map[string]*string, len(req.Inputs))
		for name, value := range req.Inputs {
			value := value
			supplied[name] = &value
		}
		supplied = MergeRunbookInputDefaults(&runbookConfig, supplied)
		if err := h.ValidateRunbookInputs(&runbookConfig, supplied); err != nil {
			return err
		}
		if req.TriggerEventDispatchID != nil {
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.InstallRunbookRun{TriggerEventDispatchID: req.TriggerEventDispatchID}).First(&run).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if run.ID == "" && req.IdempotencyKey != nil {
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.InstallRunbookRun{IdempotencyKey: req.IdempotencyKey}).First(&run).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if run.ID == "" {
			inputs := pgtype.Hstore{}
			for _, input := range runbookConfig.Inputs {
				value := input.Default
				if suppliedValue, ok := supplied[input.Name]; ok && suppliedValue != nil {
					value = *suppliedValue
				}
				inputs[input.Name] = &value
			}
			run = app.InstallRunbookRun{OrgID: installRunbook.OrgID, InstallID: installRunbook.InstallID, InstallRunbookID: installRunbook.ID, RunbookConfigID: req.RunbookConfigID, RunbookInputs: inputs, StepSelections: req.StepSelections, Status: app.InstallRunbookRunStatusQueued, TriggeredByID: req.TriggeredByID, TriggerEventDispatchID: req.TriggerEventDispatchID, IdempotencyKey: req.IdempotencyKey}
			create := tx
			switch {
			case req.TriggerEventDispatchID != nil:
				create = create.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "trigger_event_dispatch_id"}}, DoNothing: true})
			case req.IdempotencyKey != nil:
				create = create.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "idempotency_key"}}, DoNothing: true})
			}
			result := create.Create(&run)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				// Lost the insert race against a concurrent caller with the same key;
				// adopt the run they created rather than starting a second one.
				lookup := app.InstallRunbookRun{TriggerEventDispatchID: req.TriggerEventDispatchID}
				if req.TriggerEventDispatchID == nil {
					lookup = app.InstallRunbookRun{IdempotencyKey: req.IdempotencyKey}
				}
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(lookup).First(&run).Error; err != nil {
					return err
				}
			}
		}
		if run.InstallRunbookID != installRunbook.ID || run.RunbookConfigID != req.RunbookConfigID {
			return errors.New("trigger dispatch already triggered a different runbook snapshot")
		}
		metadata := map[string]string{"install_runbook_id": installRunbook.ID, "install_runbook_run_id": run.ID, "runbook_name": installRunbook.Runbook.Name, "runbook_config_id": req.RunbookConfigID, "install_id": installRunbook.InstallID}
		if run.InstallWorkflowID == nil {
			approvalOption := app.InstallApprovalOptionPrompt
			var installConfig app.InstallConfig
			err := tx.Where(app.InstallConfig{InstallID: installRunbook.InstallID}).Order("created_at DESC").First(&installConfig).Error
			if err == nil {
				approvalOption = installConfig.ApprovalOption
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if installRunbook.Install.Name != "" {
				metadata[app.WorkflowMetadataKeyOwnerName] = installRunbook.Install.Name
			}
			created, err := h.appsHelpers.CreateWorkflowWithDB(ctx, tx, installRunbook.InstallID, "installs", app.WorkflowTypeRunbookRun, metadata, false, approvalOption, req.Role)
			if err != nil {
				return err
			}
			workflow = *created
			run.InstallWorkflowID = &workflow.ID
			if err := tx.Model(&run).Update("install_workflow_id", workflow.ID).Error; err != nil {
				return err
			}
		} else if err := tx.Where(app.Workflow{ID: *run.InstallWorkflowID}).First(&workflow).Error; err != nil {
			return err
		}
		var q app.Queue
		if err := tx.Where(app.Queue{OwnerID: installRunbook.InstallID, OwnerType: "installs", Name: "install-workflows"}).First(&q).Error; err != nil {
			return err
		}
		queueID = q.ID
		dedupe := "runbook-run:" + run.ID
		var existing app.QueueSignal
		if err := tx.Where(app.QueueSignal{QueueID: q.ID, DedupeKey: &dedupe}).First(&existing).Error; err == nil {
			queueSignalID = existing.ID
			// The signal already exists (dedupe). enqueueSignal does not merge
			// callbacks onto a deduplicated row, so register ours here — unless the
			// run already finished, in which case no completion signal is coming and
			// the caller must not wait for one.
			if req.Callback.IsSet() {
				if existing.Status.Status == app.StatusQueued || existing.Status.Status == app.StatusInProgress {
					if err := appendSignalCallback(tx, existing.ID, req.Callback); err != nil {
						return err
					}
					// Re-read after the append: the handler may have completed and
					// fanned out its callbacks between the status read above and this
					// write, which would leave ours attached to a finished signal that
					// never fires. Mirrors queue/client.EnsureSignal.
					var recheck app.QueueSignal
					if err := tx.Where(app.QueueSignal{ID: existing.ID}).First(&recheck).Error; err != nil {
						return err
					}
					if recheck.Status.Status != app.StatusQueued && recheck.Status.Status != app.StatusInProgress {
						terminalStatus = string(recheck.Status.Status)
					}
				} else {
					terminalStatus = string(existing.Status.Status)
				}
			}
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		resp, err := h.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, Signal: &executeflow.Signal{WorkflowID: workflow.ID}, OwnerID: workflow.ID, OwnerType: "install_workflows", DedupeKey: &dedupe, Callback: req.Callback})
		if err != nil {
			return err
		}
		queueSignalID = resp.ID
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("trigger runbook run: %w", err)
	}
	dedupe := "runbook-run:" + run.ID
	_, _ = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: queueID, Signal: &executeflow.Signal{WorkflowID: workflow.ID}, OwnerID: workflow.ID, OwnerType: "install_workflows", DedupeKey: &dedupe, Callback: req.Callback})
	return &TriggerRunbookRunResponse{Run: &run, Workflow: &workflow, QueueSignalID: queueSignalID, TerminalStatus: terminalStatus}, nil
}

// appendSignalCallback atomically appends a callback to an in-flight queue signal.
// Raw SQL because the append is a JSONB concatenation GORM can't express; matches
// the idiom in queue/client.EnsureSignal.
func appendSignalCallback(tx *gorm.DB, queueSignalID string, cb callback.Ref) error {
	cbJSON, err := json.Marshal([]callback.Ref{cb})
	if err != nil {
		return fmt.Errorf("unable to marshal callback: %w", err)
	}
	return tx.Exec(
		`UPDATE queue_signals SET callbacks = COALESCE(callbacks, '[]'::jsonb) || ?::jsonb WHERE id = ?`,
		string(cbJSON), queueSignalID,
	).Error
}
