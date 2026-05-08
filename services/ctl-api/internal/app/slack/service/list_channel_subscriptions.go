package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListSlackChannelSubscriptions
// @Summary				List Slack channel subscriptions for the current org
// @Description			Returns the per-channel routing rules belonging to the calling org's verified Slack org links.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Success				200	{array}		app.SlackChannelSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/channel-subscriptions [GET]
func (s *service) ListChannelSubscriptions(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	subs, err := s.listChannelSubscriptions(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list slack channel subscriptions: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, subs)
}

func (s *service) listChannelSubscriptions(ctx context.Context, orgID string) ([]app.SlackChannelSubscription, error) {
	var subs []app.SlackChannelSubscription
	res := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{OrgID: orgID}).
		Order("created_at DESC").
		Find(&subs)
	if res.Error != nil {
		return nil, res.Error
	}
	return subs, nil
}
