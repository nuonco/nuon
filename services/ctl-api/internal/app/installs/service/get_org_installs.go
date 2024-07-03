package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetOrgInstalls
// @Summary	get all installs for an org
// @Description.markdown	get_org_installs.md
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Install
// @Router			/v1/installs [GET]
func (s *service) GetOrgInstalls(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	install, err := s.getOrgInstalls(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for org %s: %w", org.ID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) getOrgInstalls(ctx context.Context, orgID string) ([]app.Install, error) {
	var installs []app.Install
	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("CreatedBy").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.Org").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("InstallComponents").
		Preload("InstallComponents.InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC")
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Preload("InstallComponents.Component").
		Joins("JOIN apps ON apps.id=installs_view_v3.app_id").
		Joins("JOIN orgs ON orgs.id=apps.org_id").
		Order("created_at desc").
		Find(&installs, "installs_view_v3.org_id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org installs: %w", res.Error)
	}

	return installs, nil
}
