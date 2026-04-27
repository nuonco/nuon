package driftcheckcomponent

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const SignalType signal.SignalType = "drift_check_component"

type Signal struct {
	InstallID        string `json:"install_id"`
	ComponentID      string `json:"component_id"`
	ComponentName    string `json:"component_name"`
	ComponentBuildID string `json:"component_build_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.ComponentID == "" {
		return errors.New("component_id is required")
	}
	if s.ComponentBuildID == "" {
		return errors.New("component_build_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	deploy, err := activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
		InstallID:   s.InstallID,
		ComponentID: s.ComponentID,
		BuildID:     s.ComponentBuildID,
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
			app.WorkflowMetadataKeyWorkflowNameSuffix: s.ComponentName,
			"install_deploy_id":                       deploy.ID,
			"deploy_dependents":                       "false",
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create workflow: %w", err)
	}

	if err := activities.AwaitUpdateInstallDeployWithWorkflow(ctx, activities.UpdateInstallDeployWithWorkflowRequest{
		InstallDeployID: deploy.ID,
		WorkflowID:      wkflw.ID,
	}); err != nil {
		return fmt.Errorf("unable to update install deploy with workflow: %w", err)
	}

	return signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installsignals.Signal]{
		ID: s.InstallID,
		Signal: &installsignals.Signal{
			Type:              installsignals.OperationExecuteFlow,
			InstallWorkflowID: wkflw.ID,
		},
	})
}
