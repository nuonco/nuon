package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FailQueuedDeploysRequest struct {
	InstallID string `validate:"required"`
}

// Since releases can be running in the background, and queueing jobs, it is possible that an install can have a deploy
// queued for it, while it's being deleted.
//
// Eventually, this should be fixed with more intelligent release tooling, but for now we just mark them as error
func (a *Activities) FailQueuedDeploys(ctx context.Context, req FailQueuedDeploysRequest) error {
	var installDeploys []app.InstallDeploy

	res := a.db.WithContext(ctx).
		Joins("JOIN install_components ON install_components.id=install_component_id").
		Where("install_components.install_id = ?", req.InstallID).
		Where(app.InstallDeploy{
			Status: "queued",
		}).
		Find(&installDeploys)
	if res.Error != nil {
		return fmt.Errorf("unable to fail queued install deploys: %w", res.Error)
	}

	installDeployIDs := make([]string, 0)
	for _, installDeploy := range installDeploys {
		installDeployIDs = append(installDeployIDs, installDeploy.ID)
	}

	if len(installDeployIDs) < 1 {
		return nil
	}

	res = a.db.WithContext(ctx).
		Table("install_deploys").
		Where("id IN ?", installDeployIDs).
		Updates(app.InstallDeploy{
			Status:            "error",
			StatusDescription: "deploy was queued while the install was being deleted",
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploys: %w", res.Error)
	}

	return nil
}
