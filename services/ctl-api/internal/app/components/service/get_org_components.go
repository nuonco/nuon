package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetOrgComponents
// @Summary	get all components for an org
// @Description.markdown	get_org_components.md
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Component
// @Router			/v1/components [GET]
func (s *service) GetOrgComponents(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	component, err := s.getOrgComponents(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get components for org %s: %w", org.ID, err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) getOrgComponents(ctx context.Context, orgID string) ([]app.Component, error) {
	comps := []app.Component{}

	res := s.db.WithContext(ctx).
		Joins("JOIN apps on apps.id=components.app_id").
		Where("apps.org_id = ?", orgID).
		Order("created_at desc").
		Find(&comps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get components: %w", res.Error)
	}

	return comps, nil
}
