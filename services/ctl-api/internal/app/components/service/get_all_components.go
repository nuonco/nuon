package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetAllComponents
// @Summary	get all components for all orgs
// @Description.markdown	get_all_components.md
// @Tags			components/admin
// @Accept			json
// @Produce		json
// @Param			X-Nuon-Org-ID	header	string	true	"org ID"
// @Success		200				{array}	app.Component
// @Router			/v1/components [get]
func (s *service) GetAllComponents(ctx *gin.Context) {
	components, err := s.getAllComponents(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get components for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, components)
}

func (s *service) getAllComponents(ctx context.Context) ([]*app.Component, error) {
	var components []*app.Component
	res := s.db.WithContext(ctx).
		Order("created_at desc").
		Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all components: %w", res.Error)
	}

	return components, nil
}
