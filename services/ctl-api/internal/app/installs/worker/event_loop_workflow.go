package worker

import (
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

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

	workflow.ExecuteChildWorkflow(ctx, w.ActionWorkflowTriggers)
	return nil
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:            w.AwaitCreated,
		signals.OperationPollDependencies:   w.AwaitPollDependencies,
		signals.OperationProvision:          w.AwaitProvision,
		signals.OperationReprovisionRunner:  w.AwaitReprovisionRunner,
		signals.OperationReprovision:        w.AwaitReprovision,
		signals.OperationDelete:             w.AwaitDelete,
		signals.OperationDeprovision:        w.AwaitDeprovision,
		signals.OperationDeprovisionRunner:  w.AwaitDeprovisionRunner,
		signals.OperationForgotten:          w.AwaitForget,
		signals.OperationDeployComponents:   w.AwaitDeployComponents,
		signals.OperationTeardownComponents: w.AwaitTeardownComponents,
		signals.OperationDeploy:             w.AwaitDeploy,
		signals.OperationActionWorkflowRun:  w.AwaitActionWorkflowRun,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			w.AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers: w.handleSyncActionWorkflowTriggers,
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		// NOTE: disabled
		// StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
		// 	w.handleSyncActionWorkflowTriggers(ctx, signals.RequestSignal{
		// 		EventLoopRequest: req,
		// 	})
		// 	return nil
		// },
	}

	return l.Run(ctx, req, pendingSignals)
}
