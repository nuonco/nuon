package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	defaultQueueWorkflowIDTemplate string = "queue-%s"
)

type CreateQueueRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`
	Namespace string `validate:"required"`

	MaxInFlight int
	MaxDepth    int
}

func (c *Client) Create(ctx context.Context, req *CreateQueueRequest) (*app.Queue, error) {
	q := app.Queue{
		OwnerID:     req.OwnerID,
		OwnerType:   req.OwnerType,
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

	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	opts := tclient.StartWorkflowOptions{
		ID:        q.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":         q.ID,
			"owner-id":   q.OwnerID,
			"owner-type": q.OwnerType,
		},
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
