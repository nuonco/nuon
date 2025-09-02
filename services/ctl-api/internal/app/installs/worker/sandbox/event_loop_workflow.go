package sandbox

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationDeprovisionSandboxPlan:      AwaitDeprovisionSandboxPlan,
		signals.OperationDeprovisionSandboxApplyPlan: AwaitDeprovisionSandboxApplyPlan,

		signals.OperationReprovisionSandboxPlan:      AwaitReprovisionSandboxPlan,
		signals.OperationReprovisionSandboxApplyPlan: AwaitReprovisionSandboxApplyPlan,

		signals.OperationProvisionSandboxPlan:      AwaitProvisionSandboxPlan,
		signals.OperationProvisionSandboxApplyPlan: AwaitProvisionSandboxApplyPlan,
	}
}

func (w *Workflows) SandboxEventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := w.GetHandlers()
	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook: func(ctx workflow.Context, elr eventloop.EventLoopRequest) error {
			installSandbox, err := activities.AwaitGetInstallSandbox(ctx, activities.GetInstallSandboxRequest{
				ID: elr.ID,
			})
			if err != nil {
				return fmt.Errorf("unable to get install sandbox: %w", err)
			}

			install, err := activities.AwaitGetByInstallID(ctx, installSandbox.InstallID)
			if err != nil {
				return fmt.Errorf("unable to get install for sandbox: %w", err)
			}

			if install.AppSandboxConfig.DriftSchedule == "" {
				// no drift schedule, nothing to do.
				return nil
			}

			w.startDriftDetectionSandbox(ctx, DriftRequest{
				InstallID:     installSandbox.InstallID,
				DriftSchedule: install.AppSandboxConfig.DriftSchedule,
			})

			return nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}

type DriftRequest struct {
	InstallID     string `validate:"required" json:"install_id"`
	DriftSchedule string `validate:"required" json:"drift_schedule"`
}

func (w *Workflows) startDriftDetectionSandbox(ctx workflow.Context, req DriftRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            driftWorkflowID(req.InstallID),
		CronSchedule:          req.DriftSchedule,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.DriftCheckSandbox, req)
}

func driftWorkflowID(installID string) string {
	return "drift-check-sandbox-" + installID
}

func (w *Workflows) DriftCheckSandbox(ctx workflow.Context, req *DriftRequest) error {
	wkflw, err := activities.AwaitCreateWorkflow(ctx, activities.CreateWorkflowRequest{
		InstallID:    req.InstallID,
		WorkflowType: app.WorkflowTypeDriftRunReprovisionSandbox,
		ErrBehavior:  app.StepErrorBehaviorAbort,
		Metadata:     map[string]string{},
		PlanOnly:     true,
	})
	if err != nil {
		return fmt.Errorf("unable to create workflow: %w", err)
	}

	w.evClient.Send(ctx, req.InstallID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: wkflw.ID,
	})

	return nil
}
