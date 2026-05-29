package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetDatadogConnection
// @Summary				Get a Datadog connection
// @Description			Returns a single DatadogConnection. The connection must belong to the calling org (ABAC enforced at the DB query level — a connection from another org returns 404, not 403, to avoid existence-disclosure).
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					connection_id	path	string	true	"Connection ID"
// @Success				200		{object}	app.DatadogConnection
// @Failure				404		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections/{connection_id} [GET]
func (s *service) GetConnection(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connectionID := ctx.Param("connection_id")

	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: connectionID, OrgID: org.ID}).
		First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         fmt.Errorf("datadog connection %q not found in org %q", connectionID, org.ID),
				Description: "Datadog connection not found",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to fetch datadog connection: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, conn)
}
