package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAllComponents
// @Summary				get all components for all orgs
// @Description.markdown	get_all_components.md
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
	var components []*app.Component
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Order("created_at desc").
		Preload("Dependencies").
		Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all components: %w", res.Error)
	}

	components, err := db.HandlePaginatedResponse(ctx, components)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return components, nil
}
