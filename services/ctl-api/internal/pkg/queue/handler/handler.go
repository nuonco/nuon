package handler

import (
	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// StartHandler starts a new handler workflow for a signal. Uses TERMINATE_IF_RUNNING
// to replace any stale handler, and PARENT_CLOSE_POLICY_ABANDON so the handler
// survives when the parent queue workflow closes (gracefully or via termination).
func StartHandler(ctx workflow.Context, workflowID string, req HandlerRequest) {
	_ = (&Workflows{}).Handler
	// use this ^ for to go-to-definition jumping in your editor

	disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
		WorkflowID:            workflowID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_ABANDON,
		WaitForCancellation:   false,
	}
	disconnectedCtx = workflow.WithChildOptions(disconnectedCtx, cwo)

	workflow.ExecuteChildWorkflow(disconnectedCtx, (&Workflows{}).Handler, req)
}

// StartHandlerIfNeeded starts a handler workflow only if one is not already running.
// Uses ALLOW_DUPLICATE so a new handler is started for any previously-closed execution
// (completed, failed, cancelled, or terminated). If the handler is still running, the
// start attempt fails silently (fire-and-forget) and the existing handler continues.
// Uses PARENT_CLOSE_POLICY_ABANDON so the handler survives when the parent queue closes.
func StartHandlerIfNeeded(ctx workflow.Context, workflowID string, req HandlerRequest) {
	disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
		WorkflowID:            workflowID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_ABANDON,
		WaitForCancellation:   false,
	}
	disconnectedCtx = workflow.WithChildOptions(disconnectedCtx, cwo)

	workflow.ExecuteChildWorkflow(disconnectedCtx, (&Workflows{}).Handler, req)
}

type HandlerRequest struct {
	QueueID       string `validate:"required"`
	QueueSignalID string `validate:"required"`
}

// @temporal-gen-v2 workflow
// @task-queue "handler"
// @id-template queue-{{.QueueID}}-handler-{{.QueueSignalID}}
func (w *Workflows) Handler(ctx workflow.Context, req HandlerRequest) error {
	h := &handler{
		cfg:           w.cfg,
		v:             w.v,
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

	queueID       string
	queueSignalID string

	ready     bool
	stopped   bool
	restarted bool
	finished  bool
	canceled  bool

	// validated and executing track whether validate/execute have been performed.
	// These make the update handlers idempotent: a second validate is a no-op, and
	// a second execute waits for the in-progress execution to finish.
	validated bool
	executing bool

	// cancelable context for execution
	executingCtx    workflow.Context
	executingCancel workflow.CancelFunc

	// state that is loaded during run, but not passed between continue-as-news
	queueSignal *app.QueueSignal
	sig         signal.Signal
}
