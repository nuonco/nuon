package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
)

type SetCronEmittersEnabledRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`

	Enabled bool
	Reason  string
}

type SetCronEmittersEnabledResponse struct {
	Changed    int      `json:"changed"`
	Errors     int      `json:"errors"`
	EmitterIDs []string `json:"emitter_ids"`
}

// SetCronEmittersEnabledForOwner flips every live cron emitter on the owner's
// queues to the requested state and brings their Temporal executions in line.
//
// Disabling tears the workflows down; enabling restarts them. Only rows whose
// state actually changes are touched, so repeat calls from a health sweep that
// has already converged cost one UPDATE and no Temporal traffic.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (c *Client) SetCronEmittersEnabledForOwner(ctx context.Context, req *SetCronEmittersEnabledRequest) (*SetCronEmittersEnabledResponse, error) {
	var queueIDs []string
	if res := c.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{OwnerID: req.OwnerID, OwnerType: req.OwnerType}).
		Pluck("id", &queueIDs); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list queues for owner")
	}

	resp := &SetCronEmittersEnabledResponse{}
	if len(queueIDs) == 0 {
		return resp, nil
	}

	reason := ""
	if !req.Enabled {
		reason = req.Reason
	}

	var changed []app.QueueEmitter
	if res := c.db.WithContext(ctx).
		Model(&changed).
		Clauses(clause.Returning{}).
		Where("queue_id IN ? AND mode = ? AND deleted_at = 0 AND enabled <> ?",
			queueIDs, app.QueueEmitterModeCron, req.Enabled).
		Updates(map[string]any{
			"enabled":         req.Enabled,
			"disabled_reason": reason,
		}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter enabled state")
	}

	if len(changed) == 0 {
		return resp, nil
	}

	resp.Changed = len(changed)
	for i := range changed {
		em := &changed[i]
		resp.EmitterIDs = append(resp.EmitterIDs, em.ID)

		var err error
		if req.Enabled {
			_, err = c.RestartEmitterWorkflow(ctx, em.ID)
		} else {
			err = c.stopEmitterWorkflow(ctx, em)
		}
		if err != nil {
			resp.Errors++
			c.l.Warn("unable to reconcile emitter workflow after enabled change",
				zap.String("id", em.ID),
				zap.Bool("enabled", req.Enabled),
				zap.Error(err))
		}
	}

	c.mw.Incr("queue.emitter.enabled_changed", metrics.ToTags(map[string]string{
		"owner_type": req.OwnerType,
		"enabled":    strconv.FormatBool(req.Enabled),
		"reason":     req.Reason,
	}))

	c.l.Info("cron emitters enabled state changed",
		zap.String("owner-id", req.OwnerID),
		zap.String("owner-type", req.OwnerType),
		zap.Bool("enabled", req.Enabled),
		zap.String("reason", req.Reason),
		zap.Int("changed", resp.Changed),
	)

	return resp, nil
}

// stopEmitterWorkflow asks the emitter's parent workflow to wind down. It is
// deliberately not StopEmitter, which also marks the emitter cancelled and
// would conflate a health-driven disable with a user pause.
func (c *Client) stopEmitterWorkflow(ctx context.Context, em *app.QueueEmitter) error {
	if _, err := c.tClient.UpdateWorkflowInNamespace(ctx, em.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   em.Workflow.ID,
		UpdateName:   emitter.StopUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args:         []any{&emitter.StopRequest{}},
	}); err != nil {
		return fmt.Errorf("unable to stop emitter workflow %s: %w", em.Workflow.ID, err)
	}
	return nil
}
