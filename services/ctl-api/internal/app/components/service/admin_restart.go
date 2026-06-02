package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentrestart "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type RestartComponentRequest struct{}

// @ID						AdminRestartComponent
// @Summary				restart a component
// @Description.markdown	restart_component.md
// @Param					component_id	path	string					true	"component ID"
// @Param					req				body	RestartComponentRequest	true	"Input"
// @Tags					components/admin
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/components/{component_id}/admin-restart [POST]
func (s *service) RestartComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req RestartComponentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	q, err := s.queueClient.GetQueueByOwner(ctx, component.ID, "components")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component queue: %w", err))
		return
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   component.ID,
		OwnerType: "components",
		Signal:    &componentrestart.Signal{ComponentID: component.ID},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue component restart signal: %w", err))
		return
	}
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
