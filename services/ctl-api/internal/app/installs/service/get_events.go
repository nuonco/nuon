package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID GetInstallEvents
// @Summary	get events for an install
// @Description.markdown	 get_install_events.md
// @Param			install_id	path	string	true	"install ID"
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
// @Success		200				{array}		app.InstallEvent
// @Router			/v1/installs/{install_id}/events [GET]
func (s *service) GetInstallEvents(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installEvents, err := s.getInstallEvents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install events: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installEvents)
}

func (s *service) getInstallEvents(ctx context.Context, installID string) ([]app.InstallEvent, error) {
	var installEvents []app.InstallEvent
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Where("install_id = ?", installID).
		Order("created_at desc").
		Find(&installEvents)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get events: %w", res.Error)
	}

	return installEvents, nil
}
