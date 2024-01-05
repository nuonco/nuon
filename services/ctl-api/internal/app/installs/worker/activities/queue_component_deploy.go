package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
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

	deploy := app.InstallDeploy{
		Status:            "queued",
		StatusDescription: "waiting to be deployed to install",
		ComponentBuildID:  req.BuildID,
		Type:              deployTyp,
	}
	err := a.db.WithContext(ctx).Where(&app.InstallComponent{
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
		a.hooks.InstallDeployCreated(ctx, req.InstallID, deploy.ID)
	}

	return &deploy, nil
}
