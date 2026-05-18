package client

import (
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
)

type Client struct {
	db       *gorm.DB
	cfg      *internal.Config
	tClient  temporalclient.Client
	l        *zap.Logger
	mw       metrics.Writer
	enqueuer *enqueuer.Enqueuer
}

type Params struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	Cfg      *internal.Config
	TClient  temporalclient.Client
	L        *zap.Logger
	MW       metrics.Writer
	Enqueuer *enqueuer.Enqueuer
}

func New(params Params) *Client {
	return &Client{
		db:       params.DB,
		cfg:      params.Cfg,
		tClient:  params.TClient,
		l:        params.L,
		mw:       params.MW,
		enqueuer: params.Enqueuer,
	}
}

// queueMemo returns the standard memo map for a queue workflow.
func queueMemo(q *app.Queue) map[string]any {
	m := map[string]any{
		"type":          "queue",
		"id":            q.ID,
		"name":          q.Name,
		"owner-id":      q.OwnerID,
		"owner-type":    q.OwnerType,
		"max-in-flight": q.MaxInFlight,
		"max-depth":     q.MaxDepth,
		"idle-timeout":  time.Duration(q.IdleTimeout).String(),
	}
	if q.OrgID != nil {
		m["org-id"] = *q.OrgID
	}
	return m
}

// queueStartOperation builds a WithStartWorkflowOperation for a queue workflow.
// This is used by update-with-start calls to ensure the queue workflow is running.
func (c *Client) queueStartOperation(q *app.Queue) tclient.WithStartWorkflowOperation {
	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:                       q.Workflow.ID,
		TaskQueue:                workflows.APITaskQueue,
		Memo:                     queueMemo(q),
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	return c.tClient.NewWithStartWorkflowOperation(startOpts, "Queue", wkflowReq)
}
