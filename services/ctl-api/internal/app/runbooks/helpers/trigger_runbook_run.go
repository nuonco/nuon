package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerRunbookRunRequest struct {
	InstallRunbookID, RunbookConfigID, TriggeredByID string
	Inputs                                           map[string]string
	StepSelections                                   []app.RunbookStepSelection
	TriggerEventDispatchID                           *string
}

type TriggerRunbookRunResponse struct {
	Run           *app.InstallRunbookRun
	Workflow      *app.Workflow
	QueueSignalID string
}

func (h *Helpers) TriggerRunbookRun(ctx context.Context, req TriggerRunbookRunRequest) (*TriggerRunbookRunResponse, error) {
	var run app.InstallRunbookRun
	var workflow app.Workflow
	var queueSignalID string
	var queueID string
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
		if run.ID == "" {
			inputs := pgtype.Hstore{}
			for _, input := range runbookConfig.Inputs {
				value := input.Default
				if suppliedValue, ok := supplied[input.Name]; ok && suppliedValue != nil {
					value = *suppliedValue
				}
				inputs[input.Name] = &value
			}
			run = app.InstallRunbookRun{OrgID: installRunbook.OrgID, InstallID: installRunbook.InstallID, InstallRunbookID: installRunbook.ID, RunbookConfigID: req.RunbookConfigID, RunbookInputs: inputs, StepSelections: req.StepSelections, Status: app.InstallRunbookRunStatusQueued, TriggeredByID: req.TriggeredByID, TriggerEventDispatchID: req.TriggerEventDispatchID}
			create := tx
			if req.TriggerEventDispatchID != nil {
				create = create.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "trigger_event_dispatch_id"}}, DoNothing: true})
			}
			result := create.Create(&run)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.InstallRunbookRun{TriggerEventDispatchID: req.TriggerEventDispatchID}).First(&run).Error; err != nil {
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
			created, err := h.appsHelpers.CreateWorkflowWithDB(ctx, tx, installRunbook.InstallID, "installs", app.WorkflowTypeRunbookRun, metadata, false, approvalOption)
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
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		resp, err := h.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, Signal: executeflow.NewSignal(workflow.ID), OwnerID: workflow.ID, OwnerType: "install_workflows", DedupeKey: &dedupe})
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
	_, _ = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: queueID, Signal: executeflow.NewSignal(workflow.ID), OwnerID: workflow.ID, OwnerType: "install_workflows", DedupeKey: &dedupe})
	return &TriggerRunbookRunResponse{Run: &run, Workflow: &workflow, QueueSignalID: queueSignalID}, nil
}
