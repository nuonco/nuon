package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallComponents
// @Summary				get an installs components
// @Description.markdown	get_install_components.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
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
	paginatedComponents := []app.InstallComponent{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN components ON components.id = install_components.component_id").
		Order("created_at DESC").
		Preload("Component").
		Preload("TerraformWorkspace").
		Where("install_id = ?", installID).
		Find(&paginatedComponents)

	if tx.Error != nil {
		return nil, fmt.Errorf("unable to query install components: %w", tx.Error)
	}

	paginatedComponents, err := db.HandlePaginatedResponse(ctx, paginatedComponents)
	if err != nil {
		return nil, fmt.Errorf("failed to paginate install components: %w", err)
	}

	return paginatedComponents, nil
}
