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

// @ID						DeleteDatadogConnection
// @Summary				Delete a Datadog connection
// @Description			Soft-deletes a DatadogConnection. All associated event subscriptions and managed monitor rows cascade via the FK constraint, so a single delete cleanly removes every routing rule that depended on this tenant. Managed monitors that already exist in DD are NOT removed from DD — they remain as orphans until a user deletes them via the DD UI or via DELETE /managed-monitors/{id} before deleting the connection. The dashboard surfaces a confirmation that mentions this.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					connection_id	path	string	true	"Connection ID"
// @Success				204
// @Failure				404		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections/{connection_id} [DELETE]
func (s *service) DeleteConnection(ctx *gin.Context) {
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

	if err := s.db.WithContext(ctx).Delete(&conn).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to delete datadog connection: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
