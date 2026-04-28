package client

import (
	"context"
	"encoding/json"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

type Client struct {
	db      *gorm.DB
	cfg     *internal.Config
	tClient temporalclient.Client
	l       *zap.Logger
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Cfg     *internal.Config
	TClient temporalclient.Client
	L       *zap.Logger
}

func New(params Params) *Client {
	return &Client{
		db:      params.DB,
		cfg:     params.Cfg,
		tClient: params.TClient,
		l:       params.L,
	}
}

// queueStartOperation builds a WithStartWorkflowOperation for a queue workflow.
// This is used by update-with-start calls to ensure the queue workflow is running.
func (c *Client) queueStartOperation(q *app.Queue) tclient.WithStartWorkflowOperation {
	wkflowReq := queue.QueueWorkflowRequest{
		QueueID: q.ID,
		Version: c.cfg.Version,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:        q.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":           q.ID,
			"owner-id":     q.OwnerID,
			"owner-type":   q.OwnerType,
			"idle-timeout": time.Duration(q.IdleTimeout).String(),
		},
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	return c.tClient.NewWithStartWorkflowOperation(startOpts, "Queue", wkflowReq)
}

// updateQueueSignalMetadata merges metadata keys into a queue signal's CompositeStatus.
// This is a best-effort, fire-and-forget operation used from background goroutines.
func (c *Client) updateQueueSignalMetadata(queueSignalID string, metadata map[string]any) {
	ctx := context.Background()

	var qs app.QueueSignal
	if res := c.db.WithContext(ctx).Select("id", "status").First(&qs, "id = ?", queueSignalID); res.Error != nil {
		c.l.Warn("unable to get queue signal for metadata update",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(res.Error))
		return
	}

	if qs.Status.Metadata == nil {
		qs.Status.Metadata = make(map[string]any)
	}
	for k, v := range metadata {
		qs.Status.Metadata[k] = v
	}

	statusJSON, err := json.Marshal(qs.Status)
	if err != nil {
		c.l.Warn("unable to marshal status for metadata update",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(err))
		return
	}

	if res := c.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", queueSignalID).
		Update("status", gorm.Expr("?::jsonb", string(statusJSON))); res.Error != nil {
		c.l.Warn("unable to update queue signal metadata",
			zap.String("queue-signal-id", queueSignalID),
			zap.Error(res.Error))
	}
}
