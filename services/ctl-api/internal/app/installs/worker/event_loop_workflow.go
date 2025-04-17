package worker

import (
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
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

	workflow.ExecuteChildWorkflow(ctx, w.ActionWorkflowTriggers, sreq)
	return nil
}

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:                  w.AwaitCreated,
		signals.OperationPollDependencies:         w.AwaitPollDependencies,
		signals.OperationForget:                   w.AwaitForget,
		signals.OperationReprovisionRunner:        w.AwaitReprovisionRunner,
		signals.OperationProvisionRunner:          w.AwaitProvisionRunner,
		signals.OperationDeprovisionSandbox:       w.AwaitDeprovisionSandbox,
		signals.OperationReprovisionSandbox:       w.AwaitReprovisionSandbox,
		signals.OperationProvisionSandbox:         w.AwaitProvisionSandbox,
		signals.OperationExecuteWorkflow:          w.AwaitExecuteWorkflow,
		signals.OperationExecuteActionWorkflow:    w.AwaitExecuteActionWorkflow,
		signals.OperationExecuteDeployComponent:   w.AwaitExecuteDeployComponent,
		signals.OperationExecuteTeardownComponent: w.AwaitExecuteTeardownComponent,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			w.AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers:  w.handleSyncActionWorkflowTriggers,
		signals.OperationGenerateInstallStackVersion: w.AwaitGenerateInstallStackVersion,
		signals.OperationAwaitInstallStackVersionRun: w.AwaitInstallStackVersionRun,
		signals.OperationAwaitRunnerHealthy:          w.AwaitRunnerHealthy,

		// deprecated
		signals.OperationDeploy:             w.AwaitDeploy,
		signals.OperationDeployComponents:   w.AwaitDeployComponents,
		signals.OperationTeardownComponents: w.AwaitTeardownComponents,
		signals.OperationDeleteComponents:   w.AwaitTeardownComponents,
		signals.OperationDeprovisionRunner:  w.AwaitDeprovisionRunner,
		signals.OperationDelete:             w.AwaitDelete,
		signals.OperationProvision:          w.AwaitProvisionSandbox,
		signals.OperationReprovision:        w.AwaitReprovisionSandbox,
		signals.OperationDeprovision:        w.AwaitDeprovisionSandbox,
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
			w.handleSyncActionWorkflowTriggers(ctx, signals.RequestSignal{
				Signal: &signals.Signal{
					Type: signals.OperationSyncActionWorkflowTriggers,
				},
				EventLoopRequest: req,
			})
			return nil
		},
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			// TODO(sdboyer) remove the hardcoded response. Proper code is kept in so the import can remain
			// to avoid possibilty of subtle bugs when its enabled.
			_, _ = activities.AwaitCheckExistsByID(ctx, req.ID)
			return true, nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
