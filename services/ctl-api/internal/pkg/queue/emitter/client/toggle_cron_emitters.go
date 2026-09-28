package client

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ToggleCronEmittersRequest struct {
	OrgID         string
	InstallID     string
	Disable       bool
	Reason        string
	StopWorkflows bool
}

type ToggleCronEmittersResponse struct {
	Changed    int      `json:"changed"`
	Errors     int      `json:"errors"`
	EmitterIDs []string `json:"emitter_ids"`
}

func (c *Client) ToggleCronEmitters(ctx context.Context, req *ToggleCronEmittersRequest) (*ToggleCronEmittersResponse, error) {
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	toStatus := app.StatusInProgress
	if req.Disable {
		toStatus = app.StatusDisabled
	}

	targetStatus := app.NewCompositeStatus(ctx, toStatus)
	if req.Disable {
		targetStatus.StatusHumanDescription = req.Reason
	} else {
		targetStatus.StatusHumanDescription = "admin enabled"
	}

	var changedEmitters []app.QueueEmitter
	query := c.db.WithContext(ctx).
		Model(&changedEmitters).
		Clauses(clause.Returning{}).
		Where("mode = ? AND deleted_at = 0", app.QueueEmitterModeCron)

	if req.Disable {
		query = query.Where("status->>'status' != ?", app.StatusDisabled)
	} else {
		query = query.
			Where("status->>'status' = ?", app.StatusDisabled).
			Where("status->>'status_human_description' = ?", req.Reason)
	}

	if req.OrgID != "" {
		query = query.Where("org_id = ?", req.OrgID)
	}

	if req.InstallID != "" {
		var queueIDs []string
		if res := c.db.WithContext(ctx).
			Model(&app.Queue{}).
			Where(app.Queue{OwnerID: req.InstallID, OwnerType: "installs"}).
			Pluck("id", &queueIDs); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to list queues for install")
		}
		if len(queueIDs) == 0 {
			return &ToggleCronEmittersResponse{}, nil
		}
		query = query.Where("queue_id IN ?", queueIDs)
	}

	if res := query.Updates(map[string]any{
		"status": targetStatus,
	}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to toggle cron emitters")
	}

	resp := &ToggleCronEmittersResponse{
		Changed: len(changedEmitters),
	}
	if len(changedEmitters) == 0 {
		return resp, nil
	}

	for i := range changedEmitters {
		em := &changedEmitters[i]
		resp.EmitterIDs = append(resp.EmitterIDs, em.ID)

		if req.Disable {
			if !req.StopWorkflows {
				continue
			}
			if err := c.stopEmitterWorkflow(ctx, em); err != nil {
				resp.Errors++
				c.l.Warn("unable to stop emitter workflow after disable",
					zap.String("id", em.ID),
					zap.Error(err))
			}
			continue
		}

		if _, err := c.RestartEmitterWorkflow(ctx, em.ID); err != nil {
			resp.Errors++
			c.l.Warn("unable to restart emitter workflow after enable",
				zap.String("id", em.ID),
				zap.Error(err))
		}
	}

	c.mw.Incr("queue.emitter.status_change", metrics.ToTags(map[string]string{
		"from_status":    "any",
		"to_status":      string(toStatus),
		"scope":          toggleScope(req),
		"stop_workflows": strconv.FormatBool(req.StopWorkflows),
	}))

	c.l.Info(
		"cron emitters toggled",
		zap.String("org-id", req.OrgID),
		zap.String("install-id", req.InstallID),
		zap.Bool("disable", req.Disable),
		zap.String("reason", req.Reason),
		zap.Bool("stop-workflows", req.StopWorkflows),
		zap.Int("changed", resp.Changed),
		zap.Int("errors", resp.Errors),
	)

	return resp, nil
}

func toggleScope(req *ToggleCronEmittersRequest) string {
	switch {
	case req.InstallID != "":
		return "install"
	case req.OrgID != "":
		return "org"
	default:
		return "all"
	}
}
