package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetInstallers
// @Summary	get installers for current org
// @Description.markdown	get_installers.md
// @Tags installers
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200	{array}	app.Installer
// @Router			/v1/installers [get]
func (s *service) GetInstallers(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
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

func (s *service) getInstallers(ctx context.Context, orgID string) ([]*app.Installer, error) {
	var apps []*app.Installer
	res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Preload("Apps").
		Preload("Metadata").
		Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installers: %w", res.Error)
	}

	return apps, nil
}
