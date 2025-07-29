package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
)

const (
	maxAppConfigCount = 10
)

type GetInstallComponenetLastActivePlanResponse struct {
	ComponentDeployRunnerPlan *string
}

// @ID						GetInstallComponenetLastActivePlan
// @Summary					get an install component's previous config
// @Description.markdown	get_install_component_last_active_plan.md
// @Param					install_id	path	string	true	"install ID"
// @Param					component_id	path	string	true	"component ID"
// @Tags					runners/runner
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	GetInstallComponenetLastActivePlanResponse
// @Router					/v1/installs/{install_id}/{component_id}/last-active-plan [get]
func (s *service) GetInstallComponenetLastActivePlan(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	installComponent, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	if len(installComponent.InstallDeploys) < 2 {
		ctx.JSON(http.StatusOK, &GetInstallComponenetLastActivePlanResponse{})
		return
	}
	installDeploy := installComponent.InstallDeploys[0]

	runnerJob, err := s.helpers.GetLatestJob(ctx, &helpers.GetLatestJobRequest{
		OwnerID: installDeploy.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job for install %s: %w", installID, err))
		return
	}

	// index access should not fail since we already have checked for len 0 and 1
	ctx.JSON(http.StatusOK, &GetInstallComponenetLastActivePlanResponse{
		ComponentDeployRunnerPlan: generics.ToPtr(runnerJob.Execution.Result.Contents),
	})
}

func (s *service) getInstallComponent(ctx context.Context, installID, componentID string) (*app.InstallComponent, error) {
	installCmp := app.InstallComponent{}
	res := s.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("status = ?", app.InstallDeployStatusActive).
				Order("install_deploys.created_at DESC").
				Limit(2) // we only need the latest two deploys to get the previous config
		}).
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	return &installCmp, nil
}
