package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/installs
// Create an org's installs
// @Summary get all installs for an org
// @Schemes
// @Description get all installs for an org
// @Tags installs
// @Accept json
// @Produce json
// @Success 200 {array} app.Install
// @Router /v1/installs [GET]
func (s *service) GetOrgInstalls(ctx *gin.Context) {
	orgID := "org6h27y0rsz1oocphdb7o54zh"

	install, err := s.getOrgInstalls(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) getOrgInstalls(ctx context.Context, orgID string) ([]app.Install, error) {
	org := &app.Org{}
	res := s.db.WithContext(ctx).Preload("Apps").Preload("Installs").First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	installs := make([]app.Install, 0)
	for _, app := range org.Apps {
		installs = append(installs, app.Installs...)
	}

	return installs, nil
}
