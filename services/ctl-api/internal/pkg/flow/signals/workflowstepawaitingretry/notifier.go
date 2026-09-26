package workflowstepawaitingretry

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// executeWorkflowSignalType is executeflow.SignalType. Kept here so this
// notifier does not import the flow conductor.
const executeWorkflowSignalType qsignal.SignalType = "execute-workflow"

// Notifier dispatches outside workflow code so existing in-flight workflows
// can use the new behavior without a Temporal version gate. Notification
// failures must never fail the status update that triggered them.
type Notifier struct {
	db          *gorm.DB
	queueClient *queueclient.Client
	tclient     temporalclient.Client
	l           *zap.Logger
}

type NotifierParams struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
	TClient     temporalclient.Client
	L           *zap.Logger
}

func NewNotifier(params NotifierParams) statusactivities.FlowStatusNotifier {
	return &Notifier{
		db:          params.DB,
		queueClient: params.QueueClient,
		tclient:     params.TClient,
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

	// The flow handler withholds completion callbacks while the workflow is
	// parked, so a parent waiting on this install (an app-branch deploy group)
	// would stay in progress. Deliver the failure now. A later retry that
	// succeeds does not revive that parent.
	n.notifyParentCallbacks(ctx, l, wf.ID, errMessage)

	if n.alreadyNotified(ctx, stepID, retryIndex) {
		l.Debug("awaiting-retry notification: already enqueued for this retry index",
			zap.Int("retry_index", retryIndex))
		return
	}

	q, err := n.queueClient.GetQueueByOwnerAndName(ctx, wf.OwnerID, wf.OwnerType, queuenames.InstallSignalsQueueName)
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

// notifyParentCallbacks signals completion callbacks registered on the
// workflow's in-progress execute-workflow queue signal. Failures are logged:
// this must not fail the status update that triggered it.
func (n *Notifier) notifyParentCallbacks(ctx context.Context, l *zap.Logger, workflowID, errMessage string) {
	var signals []app.QueueSignal
	err := n.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   workflowID,
			OwnerType: (&app.Workflow{}).TableName(),
			Type:      executeWorkflowSignalType,
		}).
		Order("created_at desc").
		Find(&signals).Error
	if err != nil {
		l.Warn("awaiting-retry notification: unable to load execute-workflow signals", zap.Error(err))
		return
	}

	if errMessage == "" {
		errMessage = "install workflow failed, awaiting retry or skip"
	}
	result := callback.Result{
		Status:            string(app.StatusError),
		StatusDescription: errMessage,
	}

	for _, qs := range signals {
		if qs.Status.Status != app.StatusInProgress {
			continue
		}
		for _, ref := range completionCallbackRefs(qs) {
			if err := n.tclient.SignalWorkflowInNamespace(ctx, ref.Namespace, ref.WorkflowID, "", ref.SignalName, result); err != nil {
				l.Warn("awaiting-retry notification: unable to signal parent callback",
					zap.String("target_workflow", ref.WorkflowID),
					zap.String("signal_name", ref.SignalName),
					zap.Error(err))
				continue
			}
			l.Info("awaiting-retry notification: signaled parent callback",
				zap.String("target_workflow", ref.WorkflowID),
				zap.String("signal_name", ref.SignalName))
		}
	}
}

func completionCallbackRefs(qs app.QueueSignal) callback.Refs {
	refs := append(callback.Refs{}, qs.Callbacks...)
	if !qs.Callback.IsSet() {
		return refs
	}
	for _, ref := range refs {
		if ref.WorkflowID == qs.Callback.WorkflowID && ref.SignalName == qs.Callback.SignalName {
			return refs
		}
	}
	return append(refs, qs.Callback)
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
