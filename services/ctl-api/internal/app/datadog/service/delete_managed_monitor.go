package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
)

// @ID						DeleteDatadogManagedMonitor
// @Summary				Delete a Datadog managed monitor
// @Description			Removes the DD monitor in Datadog and soft-deletes the local row. ABAC scoped at the DB query so callers cannot delete monitors outside their org. A 404 from DD on the remote delete is treated as success (already-gone) so the local row converges with DD's state.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id		path	string	true	"Org ID"
// @Param					monitor_id	path	string	true	"Managed monitor ID"
// @Success				204
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/managed-monitors/{monitor_id} [DELETE]
func (s *service) DeleteManagedMonitor(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	monitorID := ctx.Param("monitor_id")
	if monitorID == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("monitor_id is required")))
		return
	}

	if err := s.deleteManagedMonitor(ctx, org.ID, monitorID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete datadog managed monitor: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) deleteManagedMonitor(ctx context.Context, orgID, monitorID string) error {
	var monitor app.DatadogManagedMonitor
	if err := s.db.WithContext(ctx).
		Where(app.DatadogManagedMonitor{ID: monitorID, OrgID: orgID}).
		First(&monitor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return stderr.ErrNotFound{Err: fmt.Errorf("datadog managed monitor %q not found", monitorID)}
		}
		return fmt.Errorf("lookup datadog managed monitor: %w", err)
	}

	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: monitor.ConnectionID, OrgID: orgID}).
		First(&conn).Error; err != nil {
		return fmt.Errorf("lookup parent connection: %w", err)
	}

	// Best-effort delete in DD. We tolerate 404 (already gone) and
	// non-auth 5xx to keep the local soft-delete moving forward —
	// otherwise a DD outage would prevent users from cleaning up rows
	// even when DD's own monitor was removed via the DD UI. We DO
	// surface auth errors because those usually mean the user lost
	// access entirely and silently soft-deleting would mask that.
	if conn.ApplicationKey != "" {
		baseURL := ddclient.ResolveSiteURL(conn.Site)
		if err := s.ddClient.DeleteMonitor(ctx, baseURL, conn.APIKey, conn.ApplicationKey, monitor.DDMonitorID); err != nil {
			var apiErr *ddclient.APIError
			switch {
			case errors.As(err, &apiErr) && apiErr.StatusCode == 404:
				// Already gone in DD — proceed.
				s.l.Info("datadog monitor already deleted in DD, proceeding with local soft delete",
					zap.String("monitor_id", monitor.ID),
					zap.Int64("dd_monitor_id", monitor.DDMonitorID),
					zap.String("connection_id", monitor.ConnectionID),
				)
			case errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403):
				return fmt.Errorf("datadog rejected the application key for connection %q: %w", conn.ID, err)
			default:
				return fmt.Errorf("delete datadog monitor in DD: %w", err)
			}
		}
	}

	return s.db.WithContext(ctx).Delete(&monitor).Error
}
