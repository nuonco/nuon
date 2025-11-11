package components

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationExecuteDeployComponentSyncAndPlan: AwaitExecuteDeployComponentSyncAndPlan,
		signals.OperationExecuteDeployComponentApplyPlan:   AwaitExecuteDeployComponentApplyPlan,

		signals.OperationExecuteDeployComponentSyncImage: AwaitExecuteDeployComponentSyncImage,

		signals.OperationExecuteTeardownComponentSyncAndPlan: AwaitExecuteTeardownComponentSyncAndPlan,
		signals.OperationExecuteTeardownComponentApplyPlan:   AwaitExecuteTeardownComponentApplyPlan,
	}
}

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := w.GetHandlers()
	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook: func(ctx workflow.Context, elr eventloop.EventLoopRequest) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return err
			}

			installComponent, err := activities.AwaitGetInstallComponentByID(ctx, elr.ID)
			if err != nil {
				return fmt.Errorf("unable to get component build: %w", err)
			}

			switch installComponent.Component.Type {
			case app.ComponentTypeDockerBuild, app.ComponentTypeExternalImage:
				return nil
			}

			// todo(ht): should we be fetching latest build?
			// installComponent-> Install -> app config ID.
			// installComponent-> component ID.
			// component ID + app config ID ->  component config connection id
			// component config connection id -> component build
			componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, installComponent.ComponentID)
			if err != nil {
				isNotFound := errors.Is(err, gorm.ErrRecordNotFound)
				l.Error("startup-hook-error", zap.Bool("is_not_found", isNotFound), zap.Error(err))

				// TODO(jm): we always return nil, in the case that a GormRecordNotFound was not emitted
				// or found in the error stack.
				return nil
			}

			var driftSchedule string
			for _, compCfg := range installComponent.Component.ComponentConfigs {
				if compCfg.ID == componentBuild.ComponentConfigConnectionID {
					if compCfg.DriftSchedule == "" {
						return nil
					}
					driftSchedule = compCfg.DriftSchedule
					break
				}
			}

			w.startDriftDetection(ctx, DriftRequest{
				InstallComponentID: installComponent.ID,
				InstallID:          installComponent.InstallID,
				ComponentName:      installComponent.Component.Name,
				ComponentID:        installComponent.ComponentID,
				ComponentBuildID:   componentBuild.ID,
				DriftSchedule:      driftSchedule,
			})
			return nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}

type DriftRequest struct {
	InstallComponentID string `validate:"required" json:"install_component_id"`
	InstallID          string `validate:"required" json:"install_id"`
	ComponentName      string `validate:"required" json:"component_name"`
	ComponentID        string `validate:"required" json:"component_id"`
	ComponentBuildID   string `validate:"required" json:"component_build_id"`
	DriftSchedule      string `validate:"required" json:"drift_schedule"`
}

func (w *Workflows) startDriftDetection(ctx workflow.Context, req DriftRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            driftWorkflowID(req.InstallID, req.ComponentName),
		CronSchedule:          req.DriftSchedule,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.DriftCheck, req)
}

func driftWorkflowID(installID, compName string) string {
	return "drift-check-" + installID + "-" + compName
}

func (w *Workflows) DriftCheck(ctx workflow.Context, req *DriftRequest) error {
	deploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   req.InstallID,
		ComponentID: req.ComponentID,
		BuildID:     req.ComponentBuildID,
		Type:        app.InstallDeployTypeApply,
	})
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}

	wkflw, err := activities.AwaitCreateWorkflow(ctx, activities.CreateWorkflowRequest{
		InstallID:    req.InstallID,
		WorkflowType: app.WorkflowTypeDriftRun,
		ErrBehavior:  app.StepErrorBehaviorAbort,
		PlanOnly:     true,
		Metadata: map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: req.ComponentName,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       "false",
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create workflow: %w", err)
	}

	err = activities.AwaitUpdateInstallDeployWithWorkflow(ctx, activities.UpdateInstallDeployWithWorkflowRequest{
		InstallDeployID: deploy.ID,
		WorkflowID:      wkflw.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to update install deploy with workflow: %w", err)
	}

	w.evClient.Send(ctx, req.InstallID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: wkflw.ID,
	})

	return nil
}
