package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallers
// @Summary				get installers for current org
// @Description.markdown	get_installers.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					installers
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Installer
// @Router					/v1/installers [get]
func (s *service) GetInstallers(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installers, err := s.getInstallers(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installers: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installers)
}

func (s *service) getInstallers(ctx *gin.Context, orgID string) ([]*app.Installer, error) {
	var installers []*app.Installer
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where("org_id = ?", orgID).
		Preload("Apps").
		Preload("Metadata").
		Order("created_at desc").
		Find(&installers)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installers: %w", res.Error)
	}

	installers, err := db.HandlePaginatedResponse(ctx, installers)
	if err != nil {
		return nil, fmt.Errorf("unable to get installers: %w", err)
	}

	return installers, nil
}
