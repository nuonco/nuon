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

	install, err = s.reorderInstallComponents(ctx, install)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reorder install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) reorderInstallComponents(ctx context.Context, install *app.Install) (*app.Install, error) {
	installComponentsByComponentID := make(map[string]app.InstallComponent)
	components := make([]app.Component, 0)
	for _, ic := range install.InstallComponents {
		installComponentsByComponentID[ic.ComponentID] = ic
		components = append(components, ic.Component)
	}

	reorderedCmp, err := s.appsHelpers.OrderComponentsByDep(ctx, components)
	if err != nil {
		return nil, fmt.Errorf("unable to order components by dependency: %w", err)
	}

	reorderInstallComponents := make([]app.InstallComponent, 0)
	for _, c := range reorderedCmp {
		if ic, ok := installComponentsByComponentID[c.ID]; ok {
			reorderInstallComponents = append(reorderInstallComponents, ic)
		}
	}

	install.InstallComponents = reorderInstallComponents

	return install, nil
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
				Where("status != ?", app.StatusAutoSkipped).
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

	return &install, nil
}
