package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type CreateInstallDeployRequest struct {
	InstallID     string
	ReleaseStepID string
}

func (a *Activities) CreateInstallDeploy(ctx context.Context, req CreateInstallDeployRequest) error {
	step := app.ComponentReleaseStep{}
	res := a.db.WithContext(ctx).
		Preload("ComponentRelease").
		Preload("ComponentRelease.ComponentBuild").
		Preload("ComponentRelease.ComponentBuild.ComponentConfigConnection").
		First(&step, "id = ?", req.ReleaseStepID)
	if res.Error != nil {
		return fmt.Errorf("unable to get release step: %w", res.Error)
	}

	// create deploy
	ctx = context.WithValue(ctx, "org_id", step.OrgID)
	installCmp := app.InstallComponent{
		InstallID:   req.InstallID,
		ComponentID: step.ComponentRelease.ComponentBuild.ComponentConfigConnection.ComponentID,
	}
	deploy := app.InstallDeploy{
		Status:                 "queued",
		StatusDescription:      "waiting to be deployed to install",
		ComponentBuildID:       step.ComponentRelease.ComponentBuildID,
		ComponentReleaseStepID: generics.ToPtr(req.ReleaseStepID),
	}
	err := a.db.WithContext(ctx).First(&installCmp, "install_id = ?", req.InstallID).
		Association("InstallDeploys").
		Append(&deploy)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}

	a.installHooks.InstallDeployCreated(ctx, req.InstallID, deploy.ID)
	return nil
}
