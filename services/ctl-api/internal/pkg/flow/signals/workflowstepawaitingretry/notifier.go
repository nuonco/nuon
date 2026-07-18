package workflowstepawaitingretry

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// Notifier dispatches outside workflow code so existing in-flight workflows
// can use the new behavior without a Temporal version gate. Notification
// failures must never fail the status update that triggered them.
type Notifier struct {
	db          *gorm.DB
	queueClient *queueclient.Client
	l           *zap.Logger
}

type NotifierParams struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
	L           *zap.Logger
}

func NewNotifier(params NotifierParams) statusactivities.FlowStatusNotifier {
	return &Notifier{
		db:          params.DB,
		queueClient: params.QueueClient,
		l:           params.L,
	}
}

func (n *Notifier) FlowStatusUpdated(ctx context.Context, req statusactivities.UpdateStatusRequest) {
	if req.Status.Status != app.StatusFailedPendingRetry {
		return
	}

	l := n.l.With(zap.String("workflow_id", req.ID))

	stepID, _ := req.Status.Metadata["step_id"].(string)
	if stepID == "" {
		l.Warn("awaiting-retry notification: flow status update has no step_id metadata")
		return
	}
	l = l.With(zap.String("step_id", stepID))

	var wf app.Workflow
	if err := n.db.WithContext(ctx).Where(app.Workflow{ID: req.ID}).First(&wf).Error; err != nil {
		l.Warn("awaiting-retry notification: unable to load workflow", zap.Error(err))
		return
	}
	if wf.OwnerType != "installs" {
		return
	}

	var step app.WorkflowStep
	if err := n.db.WithContext(ctx).Where(app.WorkflowStep{ID: stepID}).First(&step).Error; err != nil {
		l.Warn("awaiting-retry notification: unable to load workflow step", zap.Error(err))
		return
	}

	errMessage, _ := step.Status.Metadata["reason"].(string)
	if errMessage == "" {
		errMessage = step.Status.StatusHumanDescription
	}
	retryIndex := intFromMetadata(step.Status.Metadata, "retry_index")
	maxRetries := intFromMetadata(step.Status.Metadata, "max_retries")

	if n.alreadyNotified(ctx, stepID, retryIndex) {
		l.Debug("awaiting-retry notification: already enqueued for this retry index",
			zap.Int("retry_index", retryIndex))
		return
	}

	q, err := n.queueClient.GetQueueByOwnerAndName(ctx, wf.OwnerID, wf.OwnerType, installSignalsQueueName)
	if err != nil {
		l.Warn("awaiting-retry notification: unable to find install signals queue", zap.Error(err))
		return
	}

	_, err = n.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &Signal{
			OrgID:        wf.OrgID,
			InstallID:    wf.OwnerID,
			WorkflowID:   wf.ID,
			WorkflowType: string(wf.Type),
			StepID:       step.ID,
			StepName:     step.Name,
			ErrMessage:   errMessage,
			RetryIndex:   retryIndex,
			MaxRetries:   maxRetries,
		},
		OwnerID:   step.ID,
		OwnerType: (&app.WorkflowStep{}).TableName(),
	})
	if err != nil {
		l.Warn("awaiting-retry notification: unable to enqueue signal", zap.Error(err))
		return
	}

	l.Info("awaiting-retry notification enqueued", zap.Int("retry_index", retryIndex))
}

// alreadyNotified guards against a Temporal activity retry enqueueing the
// same step retry twice.
func (n *Notifier) alreadyNotified(ctx context.Context, stepID string, retryIndex int) bool {
	var last app.QueueSignal
	err := n.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   stepID,
			OwnerType: (&app.WorkflowStep{}).TableName(),
			Type:      SignalType,
		}).
		Order("created_at desc").
		First(&last).Error
	if err != nil {
		return false
	}

	prev, ok := last.Signal.Signal.(*Signal)
	return ok && prev.RetryIndex == retryIndex
}

func intFromMetadata(meta map[string]any, key string) int {
	switch v := meta[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}
