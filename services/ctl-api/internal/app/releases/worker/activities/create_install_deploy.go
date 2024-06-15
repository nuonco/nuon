package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
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
	componentID := step.ComponentRelease.ComponentBuild.ComponentConfigConnection.ComponentID

	// set the orgID on the context, for all writes
	ctx = middlewares.SetOrgIDContext(ctx, step.OrgID)

	// ensure that the install component exists
	var install app.Install
	res = a.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Where("component_id = ?", componentID).
				Where("install_id = ?", req.InstallID)
		}).
		First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return fmt.Errorf("unable to get install: %w", res.Error)
	}

	// if the install component does not exist, create it.
	if len(install.InstallComponents) != 1 {
		err := a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			First(&install, "id = ?", req.InstallID).
			Association("InstallComponents").
			Append(&app.InstallComponent{
				ComponentID: componentID,
			})
		if err != nil {
			return fmt.Errorf("unable to create missing install component: %w", err)
		}
	}

	// create deploy
	installCmp := app.InstallComponent{}
	deploy := app.InstallDeploy{
		CreatedByID:            step.CreatedByID,
		Status:                 "queued",
		StatusDescription:      "waiting to be deployed to install",
		ComponentBuildID:       step.ComponentRelease.ComponentBuildID,
		ComponentReleaseStepID: generics.ToPtr(req.ReleaseStepID),
		Type:                   app.InstallDeployTypeRelease,
	}
	err := a.db.WithContext(ctx).Where(&app.InstallComponent{
		InstallID:   req.InstallID,
		ComponentID: componentID,
	}).First(&installCmp).
		Association("InstallDeploys").
		Append(&deploy)

	// install was deleted, or no longer exists
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("no install component found: %w", err)
	}
	if err != nil {
		return fmt.Errorf("unable to create install deploy: %w", err)
	}

	a.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:     signals.OperationDeploy,
		DeployID: deploy.ID,
	})
	return nil
}
