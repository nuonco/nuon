package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type CreateInstallDeployRequest struct {
	InstallID   string `json:"install_id"`
	ComponentID string `json:"component_id"`
	BuildID     string `json:"build_id"`
	Teardown    bool   `json:"teardown"`
	Signal      bool   `json:"signal"`
}

// @await-gen
func (a *Activities) CreateInstallDeploy(ctx context.Context, req CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	// create deploy
	deployTyp := app.InstallDeployTypeInstall
	if req.Teardown {
		deployTyp = app.InstallDeployTypeTeardown
	}

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
		return nil, fmt.Errorf("unable to get install component: %w", err)
	}

	installDeploy := app.InstallDeploy{
		InstallComponentID: installCmp.ID,
		CreatedByID:        install.CreatedByID,
		OrgID:              install.OrgID,
		Status:             "queued",
		StatusDescription:  "waiting to be deployed to install",
		ComponentBuildID:   req.BuildID,
		Type:               deployTyp,
	}

	res = a.db.WithContext(ctx).Create(&installDeploy)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", err)
	}

	if req.Signal {
		a.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:     signals.OperationDeploy,
			DeployID: installDeploy.ID,
		})
	}

	return &installDeploy, nil
}
