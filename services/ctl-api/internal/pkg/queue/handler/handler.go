package handler

import (
	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func StartHandler(ctx workflow.Context, workflowID string, req HandlerRequest) {
	_ = (&Workflows{}).Handler
	// use this ^ for to go-to-definition jumping in your editor

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
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

	// Callback fields for the signal-based await pattern.
	// When set, the handler sends Temporal signals on completion instead of
	// relying on the caller's blocking update/activity to detect completion.
	QueueCallbackWorkflowID string `json:"queue_callback_workflow_id,omitempty"`
	QueueCallbackSignalName string `json:"queue_callback_signal_name,omitempty"`
	QueueCallbackNamespace  string `json:"queue_callback_namespace,omitempty"`

	ParentCallbackWorkflowID string `json:"parent_callback_workflow_id,omitempty"`
	ParentCallbackSignalName string `json:"parent_callback_signal_name,omitempty"`
	ParentCallbackNamespace  string `json:"parent_callback_namespace,omitempty"`
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

		queueCallbackWorkflowID: req.QueueCallbackWorkflowID,
		queueCallbackSignalName: req.QueueCallbackSignalName,
		queueCallbackNamespace:  req.QueueCallbackNamespace,

		parentCallbackWorkflowID: req.ParentCallbackWorkflowID,
		parentCallbackSignalName: req.ParentCallbackSignalName,
		parentCallbackNamespace:  req.ParentCallbackNamespace,
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

	// finishedStatus and finishedErr capture the terminal outcome so the
	// finishedHandler can return it to AwaitSignal callers without a DB round-trip.
	finishedStatus app.Status
	finishedErr    string

	// cancelable context for execution
	executingCtx    workflow.Context
	executingCancel workflow.CancelFunc

	// Callback info for signal-based await pattern.
	// When set, completion signals are sent instead of relying on blocking updates.
	queueCallbackWorkflowID string
	queueCallbackSignalName string
	queueCallbackNamespace  string

	parentCallbackWorkflowID string
	parentCallbackSignalName string
	parentCallbackNamespace  string

	// state that is loaded during run, but not passed between continue-as-news
	queueSignal *app.QueueSignal
	sig         signal.Signal
}

// setFinished marks the handler as finished with a terminal status and optional error description.
func (h *handler) setFinished(status app.Status, errDesc string) {
	h.finished = true
	h.finishedStatus = status
	h.finishedErr = errDesc
}
