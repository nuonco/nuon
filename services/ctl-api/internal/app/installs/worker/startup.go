package worker

import (
	"fmt"
	"strings"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func (w *Workflows) startup(ctx workflow.Context, req eventloop.EventLoopRequest) error {
	sreq := signals.RequestSignal{
		Signal: &signals.Signal{
			Type: signals.OperationSyncActionWorkflowTriggers,
		},
		EventLoopRequest: req,
	}
	w.handleSyncActionWorkflowTriggers(ctx, sreq)
	w.startChildren(ctx, sreq)
	return nil
}

func (w *Workflows) handleSyncActionWorkflowTriggers(ctx workflow.Context, sreq signals.RequestSignal) error {
	workflowID := sreq.WorkflowID(sreq.ID) + "-action-workflows"
	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
		WorkflowID:            workflowID,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		// WaitForCancellation:   true,
		ParentClosePolicy: enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.ActionWorkflowTriggers, sreq)
	return nil
}

func (w *Workflows) startChildren(pctx workflow.Context, sreq signals.RequestSignal) error {
	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:             "api",
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(pctx, sreq.ID)
	if err != nil && !strings.Contains(err.Error(), "record not found") {
		return err
	}

	// older installs may not have a stack
	if stack != nil {
		cwo.WorkflowID = fmt.Sprintf("%s-%s-%s", sreq.WorkflowID(sreq.ID), "stack", stack.ID)
		ctx := workflow.WithChildOptions(pctx, cwo)
		subsreq := sreq
		subsreq.ID = stack.ID
		// NOTE(sdboyer) re-using sreq here feels like a hack that we need to get away from in a proper system
		workflow.ExecuteChildWorkflow(ctx, w.subwfStack.StackEventLoop, subsreq)
	}

	{
		sandbox, err := activities.AwaitGetInstallSandboxByInstallID(pctx, sreq.ID)
		if err != nil {
			return err
		}
		cwo.WorkflowID = fmt.Sprintf("%s-%s-%s", sreq.WorkflowID(sreq.ID), "sandbox", sandbox.ID)
		subsreq := sreq
		subsreq.ID = sandbox.ID
		ctx := workflow.WithChildOptions(pctx, cwo)
		workflow.ExecuteChildWorkflow(ctx, w.subwfSandbox.SandboxEventLoop, subsreq)
	}

	componentIDs, err := activities.AwaitGetInstallComponentIDsByInstallID(pctx, sreq.ID)
	if err != nil {
		return err
	}
	for _, id := range componentIDs {
		cwo.WorkflowID = fmt.Sprintf("%s-%s-%s", sreq.WorkflowID(sreq.ID), "component", id)
		ctx := workflow.WithChildOptions(pctx, cwo)
		subsreq := sreq
		subsreq.ID = id
		workflow.ExecuteChildWorkflow(ctx, w.subwfComponents.ComponentEventLoop, subsreq)
	}

	iaws, err := activities.AwaitGetActionWorkflowsByInstallID(pctx, sreq.ID)
	if err != nil {
		return err
	}
	for _, iaw := range iaws {
		cwo.WorkflowID = fmt.Sprintf("%s-%s-%s", sreq.WorkflowID(sreq.ID), "action", iaw.ID)
		ctx := workflow.WithChildOptions(pctx, cwo)
		subsreq := sreq
		subsreq.ID = iaw.ID
		workflow.ExecuteChildWorkflow(ctx, w.subwfActions.ActionEventLoop, subsreq)
	}

	return nil
}
