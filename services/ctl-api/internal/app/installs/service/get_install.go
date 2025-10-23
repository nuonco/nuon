package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

// @ID						GetInstall
// @Summary				get an install
// @Description.markdown	get_install.md
// @Param					install_id	path	string	true	"install ID"
// @Param					include_drifted_objects	query	bool	false	"whether to include drifted objects" Default(false)
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
// @Success				200	{object}	app.Install
// @Router					/v1/installs/{install_id} [get]
func (s *service) GetInstall(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	includeDriftedObjects := ctx.DefaultQuery("include_drifted_objects", "false")
	if includeDriftedObjects == "true" {
		dos, err := s.findDriftedObjects(ctx, org.ID, install.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get drifted objects for install %s: %w", installID, err))
			return
		}
		if dos != nil {
			install.DriftedObjects = dos
		}
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) findInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.Org").
		Preload("CreatedBy").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC")).Limit(1)
		}).
		Preload("InstallComponents").
		Preload("InstallComponents.Component").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("InstallSandbox").
		Preload("InstallSandbox.TerraformWorkspace").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("install_sandbox_runs.created_at DESC").
				Limit(5)
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Preload("InstallConfig").
		Where("name = ? AND org_id = ?", installID, orgID).
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	driftedObj := make([]app.DriftedObject, 0)
	res = s.db.WithContext(ctx).
		Where("install_id = ?", install.ID).
		Find(&driftedObj)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get drifted objects: %w", res.Error)
	}
	install.DriftedObjects = driftedObj

	return &install, nil
}
