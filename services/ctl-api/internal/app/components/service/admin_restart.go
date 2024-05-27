package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
)

type RestartComponentRequest struct{}

// @ID AdminRestartComponent
// @Summary	restart an components event loop
// @Description.markdown restart_component.md
// @Param			component_id	path	string					true	"component ID"
// @Param			req				body	RestartComponentRequest	true	"Input"
// @Tags			components/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/components/{component_id}/admin-restart [POST]
func (s *service) RestartComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req RestartComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	s.evClient.Send(ctx, component.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}

func (s *service) getComponent(ctx context.Context, componentID string) (*app.Component, error) {
	component := app.Component{}
	res := s.db.WithContext(ctx).
		Where("id = ?", componentID).
		Or("name = ?", componentID).
		Preload("ComponentConfigs").
		Preload("App").
		Preload("App.Org").
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &component, nil
}
