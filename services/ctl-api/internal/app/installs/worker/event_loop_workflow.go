package worker

import (
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:            AwaitCreated,
		signals.OperationPollDependencies:   AwaitPollDependencies,
		signals.OperationForget:             AwaitForget,
		signals.OperationExecuteFlow:        AwaitExecuteFlow,
		signals.OperationRerunFlow:          AwaitRerunFlow,
		signals.OperationWorkflowApproveAll: AwaitWorkflowApproveAll,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers: w.handleSyncActionWorkflowTriggers,

		// NOTE(jm): these should be cross account to the runners namespace
		signals.OperationAwaitRunnerHealthy: w.AwaitRunnerHealthy,
		signals.OperationProvisionRunner:    AwaitProvisionRunner,
		signals.OperationReprovisionRunner:  AwaitProvisionRunner,

		// NOTE(jm): these should be child loops
		signals.OperationProvisionDNS:   AwaitProvisionDNS,
		signals.OperationDeprovisionDNS: AwaitDeprovisionDNS,
		signals.OperationSyncSecrets:    AwaitSyncSecrets,
	}
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := w.getHandlers()
	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
			return w.handleSyncActionWorkflowTriggers(ctx, signals.RequestSignal{
				Signal: &signals.Signal{
					Type: signals.OperationSyncActionWorkflowTriggers,
				},
				EventLoopRequest: req,
			})
		},
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			return activities.AwaitCheckExistsByID(ctx, req.ID)
		},
	}

	return l.Run(ctx, req, pendingSignals)
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
