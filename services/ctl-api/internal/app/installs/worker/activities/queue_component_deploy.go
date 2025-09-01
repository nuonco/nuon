package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallDeployRequest struct {
	InstallID   string                `json:"install_id"`
	ComponentID string                `json:"component_id"`
	BuildID     string                `json:"build_id"`
	Type        app.InstallDeployType `json:"type"`
	WorkflowID  string                `json:"workflow_id"`
}

// @temporal-gen activity
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
	}

	res = a.db.WithContext(ctx).Create(&installDeploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
