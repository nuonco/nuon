package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallEvents
// @Summary				get events for an install
// @Description.markdown	get_install_events.md
// @Param					install_id					path	string	true	"install ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallEvent
// @Router					/v1/installs/{install_id}/events [GET]
func (s *service) GetInstallEvents(ctx *gin.Context) {
	appID := ctx.Param("install_id")

	installEvents, err := s.getInstallEvents(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install events: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installEvents)
}

func (s *service) getInstallEvents(ctx *gin.Context, installID string) ([]app.InstallEvent, error) {
	var installEvents []app.InstallEvent
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where("install_id = ?", installID).
		Order("created_at desc").
		Find(&installEvents)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get events: %w", res.Error)
	}

	installEvents, err := db.HandlePaginatedResponse(ctx, installEvents)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installEvents, nil
}
