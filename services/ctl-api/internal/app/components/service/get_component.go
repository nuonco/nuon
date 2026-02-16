package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetComponent
// @Summary				get a component
// @Description.markdown	get_component.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Component
// @Router					/v1/components/{component_id} [get]
func (s *service) GetComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	component, err := s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component %s: %w", componentID, err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) findComponent(ctx context.Context, orgID, componentID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where(s.db.Where("id = ?", componentID).Or("name = ?", componentID)).
		Preload("ComponentConfigs").
		Preload("Dependencies").
		Preload("App").
		Preload("App.Org").
		Scopes(helpers.PreloadLatestConfig).
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &component, nil
}

func (s *service) getComponentWithParents(ctx context.Context, cmpID string) (*app.Component, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get org from context: %w", err)
	}

	parentCmp := app.Component{}
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs").
		Preload("App").
		Preload("App.Org").
		Preload("App.Org.VCSConnections").
		Scopes(helpers.PreloadLatestConfig).
		First(&parentCmp, "id = ? AND org_id = ?", cmpID, orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &parentCmp, nil
}
