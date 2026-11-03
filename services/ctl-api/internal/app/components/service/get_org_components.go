package service

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgComponents
// @Summary				get all components for an org
// @Description.markdown	get_org_components.md
// @Param 				component_ids		query	string	false	"comma-separated list of component IDs to filter by"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		app.Component
// @Router					/v1/components [GET]
func (s *service) GetOrgComponents(ctx *gin.Context) (any, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	return s.getOrgComponents(ctx, access.OrgID())
}

func (s *service) getOrgComponents(ctx *gin.Context, orgID string) ([]app.Component, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	query, err := access.Query()
	if err != nil {
		return nil, err
	}
	componentIDs := ctx.Query("component_ids")
	var componentIDsSlice []string
	if componentIDs != "" {
		componentIDsSlice = pq.StringArray(strings.Split(strings.TrimSpace(componentIDs), ","))
		for i, id := range componentIDsSlice {
			componentIDsSlice[i] = strings.TrimSpace(id)
		}
	}

	comps := []app.Component{}

	res := query.WithContext(ctx.Request.Context()).
		Scopes(scopes.WithOffsetPagination).
		Where(&app.Component{OrgID: orgID})

	if len(componentIDsSlice) > 0 {
		res = res.Where("components.id IN ?", componentIDsSlice)
	}

	res = res.Order("components.created_at DESC, components.id DESC").
		Find(&comps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get components: %w", res.Error)
	}

	comps, err = db.HandlePaginatedResponse(ctx, comps)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return comps, nil
}
