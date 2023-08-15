package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @BasePath /v1/components
// Create an org's components
// @Summary get all components for an org
// @Schemes
// @Description get all components for an org
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {array} app.Component
// @Router /v1/components [GET]
func (s *service) GetOrgComponents(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
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
	org := &app.Org{}
	res := s.db.WithContext(ctx).Preload("Apps").Preload("Apps.Components").First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	components := make([]app.Component, 0)
	for _, app := range org.Apps {
		components = append(components, app.Components...)
	}

	return components, nil
}
