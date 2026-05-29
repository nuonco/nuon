package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteDatadogEventSubscription
// @Summary				Delete a Datadog event subscription
// @Description			Soft-deletes a DatadogEventSubscription belonging to the current org. ABAC scoped at the DB query so callers cannot delete subscriptions outside their org.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Param					sub_id	path	string	true	"Subscription ID"
// @Success				204
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/event-subscriptions/{sub_id} [DELETE]
func (s *service) DeleteEventSubscription(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	subID := ctx.Param("sub_id")
	if subID == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("sub_id is required")))
		return
	}

	if err := s.deleteEventSubscription(ctx, org.ID, subID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete datadog event subscription: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) deleteEventSubscription(ctx context.Context, orgID, subID string) error {
	var sub app.DatadogEventSubscription
	res := s.db.WithContext(ctx).
		Where(app.DatadogEventSubscription{ID: subID, OrgID: orgID}).
		First(&sub)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return stderr.ErrNotFound{Err: fmt.Errorf("datadog event subscription %q not found", subID)}
	}
	if res.Error != nil {
		return res.Error
	}

	return s.db.WithContext(ctx).Delete(&sub).Error
}
