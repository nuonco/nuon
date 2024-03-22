package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetInstallEvent
// @Summary	get an install event
// @Description.markdown	get_install_event.md
// @Param			install_id	path	string	true	"install ID"
// @Param			event_id	path	string	true	"event ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.InstallEvent
// @Router			/v1/installs/{install_id}/events/{event_id} [get]
func (s *service) GetInstallEvent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	eventID := ctx.Param("event_id")

	event, err := s.getInstallEvent(ctx, eventID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (s *service) getInstallEvent(ctx context.Context, eventID, installID string) (*app.InstallEvent, error) {
	ev := app.InstallEvent{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Where(app.InstallEvent{
			ID:        eventID,
			InstallID: installID,
		}).
		First(&ev)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install event: %w", res.Error)
	}

	return &ev, nil
}
