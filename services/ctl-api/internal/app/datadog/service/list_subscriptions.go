package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListDatadogEventSubscriptions
// @Summary				List Datadog event subscriptions for the current org
// @Description			Returns the per-connection routing rules belonging to the calling org. Optional `connection_id` query parameter scopes the list to a single connection — useful for the dashboard's per-connection detail page.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					connection_id	query	string	false	"Filter by connection ID"
// @Success				200	{array}		app.DatadogEventSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/event-subscriptions [GET]
func (s *service) ListEventSubscriptions(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	subs, err := s.listEventSubscriptions(ctx, org.ID, ctx.Query("connection_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list datadog event subscriptions: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, subs)
}

// listEventSubscriptions returns subscriptions scoped to the caller's org,
// optionally narrowed to a single connection. The org filter alone is
// sufficient for ABAC — DatadogEventSubscription.OrgID is denormalized from
// the parent connection precisely so this query doesn't need a join.
func (s *service) listEventSubscriptions(ctx context.Context, orgID, connectionID string) ([]app.DatadogEventSubscription, error) {
	where := app.DatadogEventSubscription{OrgID: orgID}
	if connectionID != "" {
		where.ConnectionID = connectionID
	}

	var subs []app.DatadogEventSubscription
	if err := s.db.WithContext(ctx).
		Where(where).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

// orgConnectionExists is a small helper used by the create/update handlers
// to confirm a caller-supplied connection_id actually belongs to the
// calling org BEFORE we attempt a write. Without this, a malicious or
// confused client could plant subscriptions under another org's
// connection by guessing IDs — the unique index alone wouldn't stop them
// because it doesn't include org_id.
func (s *service) orgConnectionExists(ctx context.Context, orgID, connectionID string) error {
	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: connectionID, OrgID: orgID}).
		First(&conn).Error; err != nil {
		return stderr.ErrNotFound{
			Err:         fmt.Errorf("datadog connection %q not found in org %q", connectionID, orgID),
			Description: "Datadog connection not found",
		}
	}
	return nil
}
