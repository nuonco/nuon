package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
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
// @Success				200	{object}	CustomerInstall
// @Router					/v1/installs/{install_id} [get]
func (s *service) GetInstall(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findCustomerInstall(ctx.Request.Context(), org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, CustomerInstall{Install: install, App: InstallAppRef{ID: install.AppID, Name: install.App.Name}})
}

func (s *service) findInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	filter := s.db.Where("name = ?", installID).Or("id = ?", installID)
	if target, ok := require.CustomerTarget(ctx); ok && target.Kind == permissions.KindInstall {
		return s.findInstallAtTarget(ctx, target)
	}
	return s.loadInstallDetails(ctx, orgID, filter)
}

func (s *service) findInstallAtTarget(ctx context.Context, target permissions.ResolvedTarget) (*app.Install, error) {
	return s.loadInstallDetails(ctx, target.OrgID, s.db.Where(app.Install{ID: target.ID, OrgID: target.OrgID, AppID: target.AppID}))
}

func (s *service) loadInstallDetails(ctx context.Context, orgID string, filter *gorm.DB) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AppBranchConnections", func(db *gorm.DB) *gorm.DB {
			return db.Order("active DESC, created_at DESC, id DESC")
		}).
		Preload("AppBranchConnections.AppBranch").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
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
		Preload("AppConfig", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "app_id", "app_branch_id")
		}).
		Preload("AppConfig.RunnerConfig").
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
		Preload("InstallRoles").
		Preload("InstallRoles.AppRoleConfig").
		Where("org_id = ?", orgID).
		Where(filter).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	driftedObj := make([]app.DriftedObject, 0)
	res = s.db.WithContext(ctx).
		Where(app.DriftedObject{OrgID: orgID, InstallID: install.ID}).
		Find(&driftedObj)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get drifted objects: %w", res.Error)
	}
	install.DriftedObjects = driftedObj

	return &install, nil
}
