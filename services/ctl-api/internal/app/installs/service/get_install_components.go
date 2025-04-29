package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

// @ID						GetInstallComponents
// @Summary				get an installs components
// @Description.markdown	get_install_components.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
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
// @Success				200	{array}		app.InstallComponent
// @Router					/v1/installs/{install_id}/components [GET]
func (s *service) GetInstallComponents(ctx *gin.Context) {
	appID := ctx.Param("install_id")
	installComponents, err := s.getInstallComponents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installComponents)
}

func (s *service) getInstallComponents(ctx *gin.Context, installID string) ([]app.InstallComponent, error) {
	install := &app.Install{}
	res := s.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Order("install_components.created_at DESC")
		}).
		Preload("InstallComponents.Component").
		Preload("TerraformWorkspace").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	// WARN: we cannot limit on a slice array with a parent that is also a slice
	// gorm will apply the limit as a single query filtered by the child ids
	for ic := range install.InstallComponents {
		latestDeploy, err := s.getLatestInstallDeploy(ctx, install.InstallComponents[ic].ID)
		if err != nil {
			return nil, fmt.Errorf("unable to get latest install deploy: %w", err)
		}

		if latestDeploy != nil {
			install.InstallComponents[ic].InstallDeploys = []app.InstallDeploy{*latestDeploy}
		}
	}

	cmps, err := db.HandlePaginatedResponse(ctx, install.InstallComponents)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	install.InstallComponents = cmps

	return install.InstallComponents, nil
}

func (s *service) getLatestInstallDeploy(ctx *gin.Context, installComponentID string) (*app.InstallDeploy, error) {
	installDeploy := &app.InstallDeploy{}
	res := s.db.WithContext(ctx).
		Where("install_component_id = ?", installComponentID).
		Order("created_at DESC").
		First(&installDeploy)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return installDeploy, nil
}
