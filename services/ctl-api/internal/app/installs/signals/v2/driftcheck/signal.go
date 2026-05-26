package driftcheck

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "drift-check"

type Signal struct {
	InstallID          string `json:"install_id"`
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "drift-check",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.InstallComponentID == "" {
		return fmt.Errorf("install_component_id is required")
	}
	if s.ComponentID == "" {
		return fmt.Errorf("component_id is required")
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("install not found: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Resolve latest component build at execution time (not emitter creation time)
	componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, s.ComponentID)
	if err != nil {
		l.Warn("drift-check: unable to get latest component build, skipping",
			"component_id", s.ComponentID,
			"error", err)
		return nil
	}

	installComponent, err := activities.AwaitGetInstallComponentByID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install component: %w", err)
	}

	deploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   s.InstallID,
		ComponentID: s.ComponentID,
		BuildID:     componentBuild.ID,
		Type:        app.InstallDeployTypeApply,
	})
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}

	wkflw, err := activities.AwaitCreateWorkflow(ctx, activities.CreateWorkflowRequest{
		InstallID:    s.InstallID,
		WorkflowType: app.WorkflowTypeDriftRun,
		PlanOnly:     true,
		Metadata: map[string]string{
			app.WorkflowMetadataKeyWorkflowNameSuffix: installComponent.Component.Name,
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

	// Enqueue the flow execution signal to the install's drift workflows queue
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: helpers.InstallDriftWorkflowsQueueName,
		Signal: &executeflow.Signal{
			WorkflowID: wkflw.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue flow execution signal: %w", err)
	}

	return nil
}
