package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

// @ID						GetInstallComponent
// @Summary				get an install component
// @Description.markdown	get_install_component.md
// @Param					install_id		path	string	true	"install ID"
// @Param					component_id	path	string	true	"component ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallComponent
// @Router					/v1/installs/{install_id}/components/{component_id} [get]
func (s *service) GetInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installCmp, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  install cmp %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, installCmp)
}

func (s *service) getInstallComponent(ctx context.Context, installID, componentID string) (*app.InstallComponent, error) {
	installCmp := app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Preload("Component").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("install_deploys.created_at DESC").Limit(1)
		}).
		Preload("TerraformWorkspace").
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	// The latest deploy's composite error is derived from its newest runner job
	// rather than stored on the row, the same way getInstallDeploy does it. The
	// component page needs it to tell a stuck Helm release apart from an ordinary
	// deploy failure, and a stale mirror would keep claiming a release is stuck
	// after it had been recovered.
	if len(installCmp.InstallDeploys) > 0 {
		latest := &installCmp.InstallDeploys[0]
		compositeError, err := runnershelpers.GetLatestJobCompositeError(ctx, s.db, runnershelpers.GetLatestJobCompositeErrorRequest{
			OwnerID:   latest.ID,
			OwnerType: "install_deploys",
		})
		if err != nil {
			s.l.Warn("unable to hydrate install component deploy composite error",
				zap.String("deploy_id", latest.ID),
				zap.Error(err))
		} else if compositeError != nil {
			latest.CompositeError = compositeError
		}
	}

	var driftedObj app.DriftedObject
	res = s.db.WithContext(ctx).
		Where("install_component_id = ?", installCmp.ID).
		First(&driftedObj)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get drifted objects: %w", res.Error)
	}
	if res.Error == nil {
		installCmp.DriftedObject = driftedObj
	}

	if err := s.populateComponentEnabled(ctx, installID, []*app.InstallComponent{&installCmp}); err != nil {
		return nil, err
	}

	return &installCmp, nil
}
