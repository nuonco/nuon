package client

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/taskqueue"
)

const (
	defaultQueueWorkflowIDTemplate string = "queue-%s"
)

type CreateQueueRequest struct {
	OrgID *string

	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`
	Namespace string `validate:"required"`

	Name            string
	Metadata        pgtype.Hstore
	SkipRestartHint bool `temporaljson:"skip_restart_hint,omitempty"`

	MaxInFlight int
	MaxDepth    int
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Create(ctx context.Context, req *CreateQueueRequest) (*app.Queue, error) {
	// Capacity must be persisted before the restart hint so the next workflow
	// run cannot reload stale limits.
	var existing app.Queue
	res := c.db.WithContext(ctx).
		Where(&app.Queue{OwnerID: req.OwnerID, OwnerType: req.OwnerType, Name: req.Name}, "owner_id", "owner_type", "name").
		First(&existing)
	if res.Error == nil {
		if err := c.reconcileQueueCapacity(ctx, &existing, req); err != nil {
			return nil, err
		}
		if existing.Workflow.Namespace != req.Namespace {
			if err := c.migrateQueueNamespace(
				ctx,
				&existing,
				req.Namespace,
				taskqueue.For(req.Namespace, req.Name),
			); err != nil {
				return nil, errors.Wrap(err, "unable to migrate queue namespace")
			}

			return &existing, nil
		}
		if !req.SkipRestartHint {
			if err := c.HintRestartSingle(ctx, existing.ID); err != nil {
				c.l.Warn("unable to hint restart existing queue during idempotent create",
					zap.String("queue-id", existing.ID), zap.Error(err))
			}
		}
		return &existing, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(res.Error, "unable to find existing queue")
	}

	q := app.Queue{
		OrgID:       req.OrgID,
		OwnerID:     req.OwnerID,
		OwnerType:   req.OwnerType,
		Name:        req.Name,
		Metadata:    req.Metadata,
		MaxInFlight: req.MaxInFlight,
		MaxDepth:    req.MaxDepth,
		Workflow: signaldb.WorkflowRef{
			Namespace:  req.Namespace,
			IDTemplate: defaultQueueWorkflowIDTemplate,
			TaskQueue:  taskqueue.For(req.Namespace, req.Name),
		},
	}
	res = c.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&q)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue")
	}
	if res.RowsAffected == 0 {
		if err := c.db.WithContext(ctx).
			Where(&app.Queue{OwnerID: req.OwnerID, OwnerType: req.OwnerType, Name: req.Name}, "owner_id", "owner_type", "name").
			First(&existing).Error; err != nil {
			return nil, errors.Wrap(err, "unable to get concurrently created queue")
		}
		if err := c.reconcileQueueCapacity(ctx, &existing, req); err != nil {
			return nil, err
		}
		if !req.SkipRestartHint {
			if err := c.HintRestartSingle(ctx, existing.ID); err != nil {
				c.l.Warn("unable to hint restart concurrently created queue",
					zap.String("queue-id", existing.ID), zap.Error(err))
			}
		}
		return &existing, nil
	}

	if c.tClient == nil {
		return &q, nil
	}

	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	opts := tclient.StartWorkflowOptions{
		ID:                    q.Workflow.ID,
		TaskQueue:             q.Workflow.TaskQueue,
		Memo:                  queueMemo(&q),
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	wkflowRun, err := c.tClient.ExecuteWorkflowInNamespace(ctx,
		q.Workflow.Namespace,
		opts,
		"Queue",
		wkflowReq,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create queue workflow")
	}
	c.l.Debug("queue started",
		zap.String("namespace", q.Workflow.Namespace),
		zap.String("id", q.Workflow.ID),
		zap.String("run-id", wkflowRun.GetRunID()),
	)

	return &q, nil
}

func (c *Client) reconcileQueueCapacity(ctx context.Context, existing *app.Queue, req *CreateQueueRequest) error {
	updates := make(map[string]any, 2)
	if existing.MaxInFlight != req.MaxInFlight {
		updates["max_in_flight"] = req.MaxInFlight
	}
	if existing.MaxDepth != req.MaxDepth {
		updates["max_depth"] = req.MaxDepth
	}
	if len(updates) == 0 {
		return nil
	}
	if err := c.db.WithContext(ctx).Model(existing).Updates(updates).Error; err != nil {
		return errors.Wrap(err, "unable to reconcile queue capacity")
	}
	existing.MaxInFlight = req.MaxInFlight
	existing.MaxDepth = req.MaxDepth
	return nil
}
