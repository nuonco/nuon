package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListDatadogManagedMonitors
// @Summary				List Datadog managed monitors for the current org
// @Description			Returns monitors created via the one-click "Alert in Datadog" action. Optional `connection_id` / `target_id` query params narrow the result set — the dashboard uses target_id to show "this install already has an alert" badges without a separate roundtrip.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					connection_id	query	string	false	"Filter by connection ID"
// @Param					target_id		query	string	false	"Filter by Nuon target ID (install/component/workflow)"
// @Success				200	{array}		app.DatadogManagedMonitor
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/managed-monitors [GET]
func (s *service) ListManagedMonitors(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	monitors, err := s.listManagedMonitors(ctx, org.ID, ctx.Query("connection_id"), ctx.Query("target_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list datadog managed monitors: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, monitors)
}

func (s *service) listManagedMonitors(ctx context.Context, orgID, connectionID, targetID string) ([]app.DatadogManagedMonitor, error) {
	where := app.DatadogManagedMonitor{OrgID: orgID}
	if connectionID != "" {
		where.ConnectionID = connectionID
	}
	if targetID != "" {
		where.TargetID = targetID
	}

	var monitors []app.DatadogManagedMonitor
	if err := s.db.WithContext(ctx).
		Where(where).
		Order("created_at DESC").
		Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}
