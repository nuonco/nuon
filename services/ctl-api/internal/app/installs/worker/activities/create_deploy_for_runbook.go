package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
)

type CreateDeployForRunbookRequest struct {
	InstallID   string `validate:"required"`
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateDeployForRunbook(ctx context.Context, req CreateDeployForRunbookRequest) (*app.InstallDeploy, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, err
	}
	var ccc app.ComponentConfigConnection
	if err := a.db.WithContext(ctx).
		Where(app.ComponentConfigConnection{AppConfigID: install.AppConfigID, ComponentID: req.ComponentID}).
		First(&ccc).Error; err != nil {
		return nil, err
	}
	build, err := buildselection.Component(ctx, a.db, ccc.ID)
	if err != nil {
		return nil, err
	}
	if build == nil {
		return nil, fmt.Errorf("no deployable build for component %s in app config %s", req.ComponentID, install.AppConfigID)
	}
	return a.CreateInstallDeploy(ctx, CreateInstallDeployRequest{
		InstallID:   req.InstallID,
		ComponentID: req.ComponentID,
		BuildID:     build.ID,
		Type:        app.InstallDeployTypeApply,
	})
}
