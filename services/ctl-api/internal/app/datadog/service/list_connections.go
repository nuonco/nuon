package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListDatadogConnections
// @Summary				List Datadog connections for the current org
// @Description			Returns every DatadogConnection belonging to the calling org, including revoked entries (so the dashboard can render the "re-enter key" state). Sorted by creation time descending so the newest connection appears first.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Success				200		{array}		app.DatadogConnection
// @Failure				401		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections [GET]
func (s *service) ListConnections(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var connections []app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{OrgID: org.ID}).
		Order("created_at DESC").
		Find(&connections).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to list datadog connections: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, connections)
}
