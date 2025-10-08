package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAllComponents
// @Summary				get all components for all orgs
// @Description.markdown	get_all_components.md
// @Param 				component_ids		query	string	false	"comma-separated list of component IDs to filter by"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					components/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{array}	app.Component
// @Router					/v1/components [get]
func (s *service) GetAllComponents(ctx *gin.Context) {
	components, err := s.getAllComponents(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get components for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, components)
}

func (s *service) getAllComponents(ctx *gin.Context) ([]*app.Component, error) {
	componentIDs := ctx.Query("component_ids")
	var componentIDsSlice []string
	if componentIDs != "" {
		componentIDsSlice = pq.StringArray(strings.Split(strings.TrimSpace(componentIDs), ","))
		for i, id := range componentIDsSlice {
			componentIDsSlice[i] = strings.TrimSpace(id)
		}
	}
	var components []*app.Component
	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Order("created_at desc").
		Preload("Dependencies")

	if len(componentIDsSlice) > 0 {
		query = query.Where("id IN ?", componentIDsSlice)
	}

	res := query.Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all components: %w", res.Error)
	}

	components, err := db.HandlePaginatedResponse(ctx, components)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return components, nil
}
