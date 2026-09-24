package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ToggleCronEmittersForOwnerRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`

	FromStatus app.Status `validate:"required"`
	ToStatus   app.Status `validate:"required"`
	Reason     string
}

type ToggleCronEmittersForOwnerResponse struct {
	Changed    int      `json:"changed"`
	Errors     int      `json:"errors"`
	EmitterIDs []string `json:"emitter_ids"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (c *Client) ToggleCronEmittersForOwner(ctx context.Context, req *ToggleCronEmittersForOwnerRequest) (*ToggleCronEmittersForOwnerResponse, error) {
	switch req.ToStatus {
	case app.StatusDisabled, app.StatusInProgress:
	default:
		return nil, errors.Errorf("unsupported target emitter status %q", req.ToStatus)
	}

	var queueIDs []string
	if res := c.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{OwnerID: req.OwnerID, OwnerType: req.OwnerType}).
		Pluck("id", &queueIDs); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list queues for owner")
	}

	resp := &ToggleCronEmittersForOwnerResponse{}
	if len(queueIDs) == 0 {
		return resp, nil
	}

	targetStatus := app.NewCompositeStatus(ctx, req.ToStatus)
	targetStatus.StatusHumanDescription = req.Reason

	var changedEmitters []app.QueueEmitter
	if res := c.db.WithContext(ctx).
		Model(&changedEmitters).
		Clauses(clause.Returning{}).
		Where("queue_id IN ? AND mode = ? AND deleted_at = 0", queueIDs, app.QueueEmitterModeCron).
		Where("status->>'status' = ?", req.FromStatus).
		Updates(map[string]any{
			"status": targetStatus,
		}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter status")
	}

	if len(changedEmitters) == 0 {
		return resp, nil
	}

	resp.Changed = len(changedEmitters)
	var reconcileErr error
	for i := range changedEmitters {
		em := &changedEmitters[i]
		resp.EmitterIDs = append(resp.EmitterIDs, em.ID)

		var err error
		if req.ToStatus == app.StatusInProgress {
			_, err = c.RestartEmitterWorkflow(ctx, em.ID)
		} else {
			err = c.stopEmitterWorkflow(ctx, em)
		}
		if err != nil {
			resp.Errors++
			previousStatus := app.NewCompositeStatus(ctx, req.FromStatus)
			if rollbackErr := c.db.WithContext(ctx).Model(em).Update("status", previousStatus).Error; rollbackErr != nil {
				c.l.Error("unable to roll back emitter status after reconciliation failure",
					zap.String("id", em.ID),
					zap.Error(rollbackErr))
			}
			if reconcileErr == nil {
				reconcileErr = errors.Wrapf(err, "unable to reconcile emitter %s", em.ID)
			}
			c.l.Warn("unable to reconcile emitter workflow after status change",
				zap.String("id", em.ID),
				zap.String("from-status", string(req.FromStatus)),
				zap.String("to-status", string(req.ToStatus)),
				zap.Error(err))
		}
	}

	c.mw.Incr("queue.emitter.status_change", metrics.ToTags(map[string]string{
		"owner_type":  req.OwnerType,
		"from_status": string(req.FromStatus),
		"to_status":   string(req.ToStatus),
	}))

	c.l.Info("cron emitter status changed",
		zap.String("owner-id", req.OwnerID),
		zap.String("owner-type", req.OwnerType),
		zap.String("from-status", string(req.FromStatus)),
		zap.String("to-status", string(req.ToStatus)),
		zap.String("reason", req.Reason),
		zap.Int("changed", resp.Changed),
	)

	return resp, reconcileErr
}
