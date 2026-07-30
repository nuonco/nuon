package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RegisterEventRunbookWaiterRequest struct {
	InstallID, WorkflowID, WorkflowStepID, QueueSignalID, TriggerID string
	EventTypes                                                      []string
	Filters                                                         []app.TriggerFilter
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) RegisterEventRunbookWaiter(ctx context.Context, req RegisterEventRunbookWaiterRequest) (*app.EventRunbookWaiter, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: req.InstallID}).First(&install).Error; err != nil {
		return nil, err
	}
	enabled, err := a.features.OrgHasFeature(ctx, install.OrgID, app.OrgFeatureTriggers)
	if err != nil || !enabled {
		return nil, fmt.Errorf("triggers feature is not enabled")
	}
	var w app.EventRunbookWaiter
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trigger app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: req.TriggerID, OrgID: install.OrgID}).First(&trigger).Error; err != nil {
			return err
		}
		w = app.EventRunbookWaiter{OrgID: install.OrgID, AppID: install.AppID, InstallID: install.ID, WorkflowID: req.WorkflowID, WorkflowStepID: req.WorkflowStepID, QueueSignalID: req.QueueSignalID, TriggerID: trigger.ID, EventTypes: req.EventTypes, Filters: req.Filters, Status: app.EventRunbookWaiterStatusActive, ActivatedAt: time.Now().UTC()}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "workflow_step_id"}}, DoNothing: true}).Create(&w).Error
	})
	if err != nil {
		return nil, err
	}
	if err := a.db.WithContext(ctx).Where(app.EventRunbookWaiter{WorkflowStepID: req.WorkflowStepID, OrgID: install.OrgID}).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

type FinishEventRunbookWaiterRequest struct {
	WorkflowStepID string
	InstallID      string
	Status         app.EventRunbookWaiterStatus
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) FinishEventRunbookWaiter(ctx context.Context, req FinishEventRunbookWaiterRequest) (app.EventRunbookWaiterStatus, error) {
	if req.Status != app.EventRunbookWaiterStatusExpired && req.Status != app.EventRunbookWaiterStatusCancelled {
		return "", fmt.Errorf("invalid terminal waiter status %q", req.Status)
	}
	now := time.Now().UTC()
	values := map[string]any{"status": req.Status}
	if req.Status == app.EventRunbookWaiterStatusExpired {
		values["expired_at"] = now
	} else {
		values["cancelled_at"] = now
	}
	if err := a.db.WithContext(ctx).Model(&app.EventRunbookWaiter{}).Where(app.EventRunbookWaiter{WorkflowStepID: req.WorkflowStepID, InstallID: req.InstallID, Status: app.EventRunbookWaiterStatusActive}).Updates(values).Error; err != nil {
		return "", err
	}
	var waiter app.EventRunbookWaiter
	if err := a.db.WithContext(ctx).Where(app.EventRunbookWaiter{WorkflowStepID: req.WorkflowStepID, InstallID: req.InstallID}).First(&waiter).Error; err != nil {
		return "", err
	}
	return waiter.Status, nil
}

type RenderRunbookActionEventOutputsRequest struct {
	ActionWorkflowRunID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) RenderRunbookActionEventOutputs(ctx context.Context, req RenderRunbookActionEventOutputsRequest) error {
	if err := a.v.Struct(req); err != nil {
		return err
	}

	var run app.InstallActionWorkflowRun
	if err := a.db.WithContext(ctx).Preload("Steps").Where(app.InstallActionWorkflowRun{ID: req.ActionWorkflowRunID}).First(&run).Error; err != nil {
		return err
	}
	if run.InstallWorkflowID == nil {
		return nil
	}
	if !runbookActionNeedsRuntimeRendering(&run) {
		return nil
	}

	data, err := a.runbookRuntimeData(ctx, *run.InstallWorkflowID)
	if err != nil {
		return err
	}

	renderedEnv, err := renderRunbookHstore(run.RunEnvVars, data)
	if err != nil {
		return fmt.Errorf("render runbook action environment: %w", err)
	}
	renderedRole, err := renderRunbookRuntimeField(run.Role, data)
	if err != nil {
		return fmt.Errorf("render runbook action role: %w", err)
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&app.InstallActionWorkflowRun{}).
			Where(app.InstallActionWorkflowRun{ID: run.ID}).
			Updates(map[string]any{"run_env_vars": renderedEnv, "role": renderedRole}).Error; err != nil {
			return err
		}

		for i := range run.Steps {
			step := &run.Steps[i]
			if step.AdHocConfig == nil {
				continue
			}
			cfg := *step.AdHocConfig
			cfg.EnvVars, err = renderRunbookHstore(cfg.EnvVars, data)
			if err != nil {
				return fmt.Errorf("render runbook action step environment: %w", err)
			}
			cfg.Command, err = renderRunbookRuntimeField(cfg.Command, data)
			if err != nil {
				return fmt.Errorf("render runbook action command: %w", err)
			}
			cfg.InlineContents, err = renderRunbookRuntimeField(cfg.InlineContents, data)
			if err != nil {
				return fmt.Errorf("render runbook action contents: %w", err)
			}
			if err := tx.Model(&app.InstallActionWorkflowRunStep{}).
				Where(app.InstallActionWorkflowRunStep{ID: step.ID}).
				Update("ad_hoc_config", &cfg).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Activities) runbookRuntimeData(ctx context.Context, workflowID string) (map[string]any, error) {
	var runbookRun app.InstallRunbookRun
	if err := a.db.WithContext(ctx).Where(app.InstallRunbookRun{InstallWorkflowID: &workflowID}).First(&runbookRun).Error; err != nil {
		return nil, err
	}

	inputs := make(map[string]string, len(runbookRun.RunbookInputs))
	for key, value := range runbookRun.RunbookInputs {
		if value != nil {
			inputs[key] = *value
		}
	}

	var waiters []app.EventRunbookWaiter
	if err := a.db.WithContext(ctx).
		Where(app.EventRunbookWaiter{WorkflowID: workflowID, Status: app.EventRunbookWaiterStatusMatched}).
		Find(&waiters).Error; err != nil {
		return nil, err
	}

	outputs := make(map[string]any, len(waiters))
	for i := range waiters {
		waiter := &waiters[i]
		if waiter.MatchedEventID == nil {
			continue
		}

		var step app.WorkflowStep
		if err := a.db.WithContext(ctx).Where(app.WorkflowStep{ID: waiter.WorkflowStepID, OrgID: waiter.OrgID}).First(&step).Error; err != nil {
			return nil, err
		}
		var event app.TriggerEvent
		if err := a.db.WithContext(ctx).
			Preload("Trigger", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).
			Where(app.TriggerEvent{ID: *waiter.MatchedEventID, OrgID: waiter.OrgID}).
			First(&event).Error; err != nil {
			return nil, err
		}

		eventOutput, err := runbookEventOutput(&event)
		if err != nil {
			return nil, fmt.Errorf("decode matched event %s payload: %w", event.ID, err)
		}
		if _, exists := outputs[step.Name]; exists {
			return nil, fmt.Errorf("multiple wait_for_event steps named %q", step.Name)
		}
		outputs[step.Name] = map[string]any{
			"event": eventOutput,
		}
	}

	return map[string]any{"runbook_inputs": inputs, "runbook_outputs": outputs}, nil
}

func runbookActionNeedsRuntimeRendering(run *app.InstallActionWorkflowRun) bool {
	containsRuntimeReference := func(value string) bool {
		return strings.Contains(value, "runbook_inputs") || strings.Contains(value, "runbook_outputs")
	}
	if containsRuntimeReference(run.Role) {
		return true
	}
	for _, value := range run.RunEnvVars {
		if value != nil && containsRuntimeReference(*value) {
			return true
		}
	}
	for i := range run.Steps {
		cfg := run.Steps[i].AdHocConfig
		if cfg == nil {
			continue
		}
		if containsRuntimeReference(cfg.Command) || containsRuntimeReference(cfg.InlineContents) {
			return true
		}
		for _, value := range cfg.EnvVars {
			if value != nil && containsRuntimeReference(*value) {
				return true
			}
		}
	}
	return false
}

func runbookEventOutput(event *app.TriggerEvent) (map[string]any, error) {
	var payload any
	decoder := json.NewDecoder(bytes.NewReader(event.Payload))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	var occurredAt any
	if event.OccurredAt != nil {
		occurredAt = event.OccurredAt.UTC().Format(time.RFC3339Nano)
	}
	return map[string]any{
		"id":          event.ID,
		"external_id": event.ExternalID,
		"source":      event.Source,
		"type":        event.EventType,
		"occurred_at": occurredAt,
		"received_at": event.ReceivedAt.UTC().Format(time.RFC3339Nano),
		"trigger": map[string]any{
			"id":   event.TriggerID,
			"name": event.Trigger.Name,
		},
		"payload": payload,
	}, nil
}

func renderRunbookHstore(values pgtype.Hstore, data map[string]any) (pgtype.Hstore, error) {
	rendered := make(pgtype.Hstore, len(values))
	for key, value := range values {
		if value == nil {
			rendered[key] = nil
			continue
		}
		result, err := renderRunbookRuntimeField(*value, data)
		if err != nil {
			return nil, err
		}
		rendered[key] = &result
	}
	return rendered, nil
}
