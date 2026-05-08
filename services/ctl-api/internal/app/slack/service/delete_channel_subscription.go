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

// @ID						DeleteSlackChannelSubscription
// @Summary				Delete a Slack channel subscription
// @Description			Soft-deletes a SlackChannelSubscription belonging to the current org. ABAC scoped at the DB query so callers cannot delete subscriptions outside their org.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Param					sub_id	path	string	true	"Slack channel subscription ID"
// @Success				204
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/channel-subscriptions/{sub_id} [DELETE]
func (s *service) DeleteChannelSubscription(ctx *gin.Context) {
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

	if err := s.deleteChannelSubscription(ctx, org.ID, subID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete slack channel subscription: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) deleteChannelSubscription(ctx context.Context, orgID, subID string) error {
	var sub app.SlackChannelSubscription
	res := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{
			ID:    subID,
			OrgID: orgID,
		}).
		First(&sub)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return stderr.ErrNotFound{Err: fmt.Errorf("slack channel subscription %q not found", subID)}
	}
	if res.Error != nil {
		return res.Error
	}

	if err := s.db.WithContext(ctx).Delete(&sub).Error; err != nil {
		return err
	}
	return nil
}
