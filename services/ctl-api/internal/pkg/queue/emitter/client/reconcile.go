package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// ReconcileActionCronEmitterRequest describes one action workflow cron emitter to reconcile.
type ReconcileActionCronEmitterRequest struct {
	InstallID               string
	QueueID                 string
	InstallActionWorkflowID string
	ActionWorkflowID        string
	CronSchedule            string // empty means "no cron, remove if exists"
	ExistingEmitterID       string // empty means no existing emitter
	ExistingCronSchedule    string // schedule of existing emitter (if any)
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ReconcileActionCronEmitter(ctx context.Context, req *ReconcileActionCronEmitterRequest) error {
	emitterName := fmt.Sprintf("action-cron-%s", req.InstallActionWorkflowID)

	// No cron needed -- remove if exists
	if req.CronSchedule == "" {
		if req.ExistingEmitterID != "" {
			c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
		}
		return nil
	}

	// Emitter exists with correct schedule -- nothing to do
	if req.ExistingEmitterID != "" && req.ExistingCronSchedule == req.CronSchedule {
		return nil
	}

	// Schedule changed or new -- delete old and recreate
	if req.ExistingEmitterID != "" {
		c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
	}

	if _, err := c.CreateEmitter(ctx, &CreateEmitterRequest{
		QueueID:      req.QueueID,
		Name:         emitterName,
		Description:  fmt.Sprintf("Cron trigger for action workflow %s", req.ActionWorkflowID),
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: req.CronSchedule,
		SignalType:   "execute_action_workflow_cron",
		SignalTemplate: signal.NewRaw("execute_action_workflow_cron", map[string]any{
			"install_id":                 req.InstallID,
			"install_action_workflow_id": req.InstallActionWorkflowID,
		}),
	}); err != nil {
		c.l.Warn("unable to create action cron emitter",
			zap.String("install_id", req.InstallID),
			zap.String("iaw_id", req.InstallActionWorkflowID),
			zap.Error(err),
		)
	}

	return nil
}

// ReconcileComponentDriftEmitterRequest describes one component drift emitter to reconcile.
type ReconcileComponentDriftEmitterRequest struct {
	InstallID          string
	QueueID            string
	InstallComponentID string
	ComponentID        string
	ComponentName      string
	ComponentBuildID   string
	DriftSchedule      string // empty means "no drift, remove if exists"
	ExistingEmitterID  string
	ExistingSchedule   string
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ReconcileComponentDriftEmitter(ctx context.Context, req *ReconcileComponentDriftEmitterRequest) error {
	emitterName := fmt.Sprintf("drift-component-%s", req.InstallComponentID)

	if req.DriftSchedule == "" {
		if req.ExistingEmitterID != "" {
			c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
		}
		return nil
	}

	if req.ExistingEmitterID != "" && req.ExistingSchedule == req.DriftSchedule {
		return nil
	}

	if req.ExistingEmitterID != "" {
		c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
	}

	if _, err := c.CreateEmitter(ctx, &CreateEmitterRequest{
		QueueID:      req.QueueID,
		Name:         emitterName,
		Description:  fmt.Sprintf("Drift detection for component %s", req.ComponentName),
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: req.DriftSchedule,
		SignalType:   "drift_check_component",
		SignalTemplate: signal.NewRaw("drift_check_component", map[string]any{
			"install_id":         req.InstallID,
			"component_id":       req.ComponentID,
			"component_name":     req.ComponentName,
			"component_build_id": req.ComponentBuildID,
		}),
	}); err != nil {
		c.l.Warn("unable to create drift component emitter",
			zap.String("install_id", req.InstallID),
			zap.String("component_id", req.ComponentID),
			zap.Error(err),
		)
	}

	return nil
}

// ReconcileSandboxDriftEmitterRequest describes the sandbox drift emitter to reconcile.
type ReconcileSandboxDriftEmitterRequest struct {
	InstallID         string
	QueueID           string
	DriftSchedule     string // empty means "no drift, remove if exists"
	ExistingEmitterID string
	ExistingSchedule  string
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) ReconcileSandboxDriftEmitter(ctx context.Context, req *ReconcileSandboxDriftEmitterRequest) error {
	emitterName := fmt.Sprintf("drift-sandbox-%s", req.InstallID)

	if req.DriftSchedule == "" {
		if req.ExistingEmitterID != "" {
			c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
		}
		return nil
	}

	if req.ExistingEmitterID != "" && req.ExistingSchedule == req.DriftSchedule {
		return nil
	}

	if req.ExistingEmitterID != "" {
		c.stopAndDeleteEmitter(ctx, req.ExistingEmitterID)
	}

	if _, err := c.CreateEmitter(ctx, &CreateEmitterRequest{
		QueueID:      req.QueueID,
		Name:         emitterName,
		Description:  "Drift detection for sandbox",
		Mode:         app.QueueEmitterModeCron,
		CronSchedule: req.DriftSchedule,
		SignalType:   "drift_check_sandbox",
		SignalTemplate: signal.NewRaw("drift_check_sandbox", map[string]any{
			"install_id": req.InstallID,
		}),
	}); err != nil {
		c.l.Warn("unable to create drift sandbox emitter",
			zap.String("install_id", req.InstallID),
			zap.Error(err),
		)
	}

	return nil
}

func (c *Client) stopAndDeleteEmitter(ctx context.Context, emitterID string) {
	if _, err := c.StopEmitter(ctx, emitterID); err != nil {
		c.l.Warn("unable to stop emitter during reconcile", zap.String("emitter_id", emitterID), zap.Error(err))
	}
	if err := c.DeleteEmitter(ctx, emitterID); err != nil {
		c.l.Warn("unable to delete emitter during reconcile", zap.String("emitter_id", emitterID), zap.Error(err))
	}
}
