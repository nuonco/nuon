package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListSlackOrgLinks
// @Summary				List Slack workspace bindings for the current org
// @Description			Returns the verified SlackOrgLink rows belonging to the calling org. Each row carries the link_id used by the channel-subscription create endpoint.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Success				200	{array}		app.SlackOrgLink
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/org-links [GET]
func (s *service) ListOrgLinks(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	links, err := s.listOrgLinks(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list slack org links: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, links)
}

// listOrgLinks returns the verified SlackOrgLink rows for orgID. Filtered to
// verified status so the dashboard never offers a revoked link as a routing
// target.
func (s *service) listOrgLinks(ctx context.Context, orgID string) ([]app.SlackOrgLink, error) {
	var links []app.SlackOrgLink
	res := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			OrgID:  orgID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		Order("created_at DESC").
		Find(&links)
	if res.Error != nil {
		return nil, res.Error
	}
	return links, nil
}
