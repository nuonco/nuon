package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

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

func (a *Activities) CreateInstallDeploy(ctx context.Context, req CreateInstallDeployRequest) (*app.InstallDeploy, error) {
	// create deploy
	installCmp := app.InstallComponent{}
	deployTyp := app.InstallDeployTypeInstall
	if req.Teardown {
		deployTyp = app.InstallDeployTypeTeardown
	}

	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, err
	}

	deploy := app.InstallDeploy{
		CreatedByID:       install.CreatedByID,
		OrgID:             install.OrgID,
		Status:            "queued",
		StatusDescription: "waiting to be deployed to install",
		ComponentBuildID:  req.BuildID,
		Type:              deployTyp,
	}
	err = a.db.WithContext(ctx).Where(&app.InstallComponent{
		InstallID:   req.InstallID,
		ComponentID: req.ComponentID,
	}).First(&installCmp).
		Association("InstallDeploys").
		Append(&deploy)

	// install was deleted, or no longer exists
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no install component found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", err)
	}

	if req.Signal {
		a.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:     signals.OperationDeploy,
			DeployID: deploy.ID,
		})
	}

	return &deploy, nil
}
