package handler

import (
	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func StartHandler(ctx workflow.Context, workflowID string, req HandlerRequest) {
	_ = (&Workflows{}).Handler

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             workflow.GetInfo(ctx).TaskQueueName,
		WorkflowID:            workflowID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_ABANDON,
		WaitForCancellation:   false,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, (&Workflows{}).Handler, req)
}

type HandlerRequest struct {
	QueueID       string `validate:"required"`
	QueueSignalID string `validate:"required"`
}

// @temporal-gen-v2 workflow
// @task-queue "handler"
// @id-template queue-{{.QueueID}}-handler-{{.QueueSignalID}}
// @memo type queue-handler
func (w *Workflows) Handler(ctx workflow.Context, req HandlerRequest) error {
	h := &handler{
		cfg:           w.cfg,
		v:             w.v,
		mw:            w.mw,
		queueSignalID: req.QueueSignalID,
		queueID:       req.QueueID,
	}

	finished, err := h.run(ctx)
	if err != nil {
		return err
	}
	if !finished {
		return workflow.NewContinueAsNewError(ctx, w.Handler, req)
	}

	return nil
}

type handler struct {
	cfg *internal.Config
	v   *validator.Validate
	mw  metrics.Writer

	queueID       string
	queueSignalID string

	ready     bool
	stopped   bool
	restarted bool
	finished  bool
	canceled  bool

	autoRewarmDeclined bool
	autoRewarmStarted  bool

	validating     bool
	executing      bool
	finishedStatus app.Status
	finishedErr    string

	executingCtx    workflow.Context
	executingCancel workflow.CancelFunc

	callbacks callback.Refs

	queueSignal     *app.QueueSignal
	sig             signal.Signal
	signalLogStream *app.LogStream
}

func (h *handler) setFinished(status app.Status, errDesc string) {
	h.finished = true
	h.finishedStatus = status
	h.finishedErr = errDesc
}
