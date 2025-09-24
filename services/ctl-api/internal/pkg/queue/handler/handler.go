package handler

import (
	"github.com/go-playground/validator/v10"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

func StartHandler(ctx workflow.Context, workflowID string, req HandlerRequest) {
	_ = (&Workflows{}).Handler
	// use this ^ for to go-to-definition jumping in your editor

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
		WorkflowID:            workflowID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WaitForCancellation:   false,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, (&Workflows{}).Handler, req)
}

type HandlerRequest struct {
	QueueID       string `validate:"required"`
	QueueSignalID string `validate:"required"`
}

// @temporal-gen workflow
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

	// state that is loaded during run, but not passed between continue-as-news
	queueSignal *app.QueueSignal
	sig         signal.Signal
}
