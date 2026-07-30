package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
)

type CreateInstallDeployRequest struct {
	InstallID   string                `json:"install_id" validate:"required"`
	ComponentID string                `json:"component_id" validate:"required"`
	BuildID     string                `json:"build_id" validate:"required"`
	Type        app.InstallDeployType `json:"type" validate:"required"`
	WorkflowID  string                `json:"workflow_id,omitempty"`
	Role        string                `json:"role,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateInstallDeploy(ctx context.Context, req CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	// create deploy
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, err
	}

	installCmp := app.InstallComponent{}
	res := a.db.WithContext(ctx).Where(&app.InstallComponent{
		InstallID:   req.InstallID,
		ComponentID: req.ComponentID,
	}).First(&installCmp)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	var installWorkflowID *string
	if req.WorkflowID != "" {
		installWorkflowID = generics.ToPtr(req.WorkflowID)
	}
	installDeploy := app.InstallDeploy{
		InstallComponentID: installCmp.ID,
		OrgID:              install.OrgID,
		Status:             "queued",
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   req.BuildID,
		Type:               req.Type,
		InstallWorkflowID:  installWorkflowID,
		Role:               req.Role,
	}

	res = a.db.WithContext(ctx).Create(&installDeploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	a.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallDeploy,
		Message:     "install deploy created",
		Outcome:     audit.OutcomeStarted,
		InstallID:   req.InstallID,
		ComponentID: req.ComponentID,
		SubjectID:   installDeploy.ID,
		SubjectType: "install_deploys",
		Attrs: map[string]string{
			"deploy.id":            installDeploy.ID,
			"deploy.type":          string(installDeploy.Type),
			"component_build.id":   req.BuildID,
			"install_component.id": installCmp.ID,
			"flow_workflow.id":     req.WorkflowID,
		},
	})

	return &installDeploy, nil
}
