package client

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	defaultQueueWorkflowIDTemplate string = "queue-%s"
)

type CreateQueueRequest struct {
	OrgID *string

	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`
	Namespace string `validate:"required"`

	Name     string
	Metadata pgtype.Hstore

	MaxInFlight int
	MaxDepth    int
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) Create(ctx context.Context, req *CreateQueueRequest) (*app.Queue, error) {
	// Idempotent: if a queue with the same owner + name already exists,
	// restart its workflow and return the existing record.
	var existing app.Queue
	if res := c.db.WithContext(ctx).
		Where(app.Queue{OwnerID: req.OwnerID, Name: req.Name}).
		First(&existing); res.Error == nil {
		if err := c.HintRestartSingle(ctx, existing.ID); err != nil {
			c.l.Warn("unable to hint restart existing queue during idempotent create",
				zap.String("queue-id", existing.ID), zap.Error(err))
		}
		return &existing, nil
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
		},
	}
	if res := c.db.WithContext(ctx).Create(&q); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue")
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
		TaskQueue:             workflows.APITaskQueue,
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
