package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

// defaultChannelsPageLimit is the default Slack page size when the caller
// doesn't specify one. Slack accepts up to 1000 but recommends ≤ 200.
const defaultChannelsPageLimit = 100

// ListChannelsResponse mirrors the subset of conversations.list surfaced to
// the dashboard.
type ListChannelsResponse struct {
	Channels   []slackclient.Conversation `json:"channels"`
	NextCursor string                     `json:"next_cursor,omitempty"`
}

// @ID						ListSlackChannels
// @Summary				List channels visible to a Slack installation
// @Description			Calls Slack's conversations.list using the installation's bot token and returns the page of channels the bot can see. The installation must belong to a verified org link for the calling org.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					installation_id	path	string	true	"Slack installation ID"
// @Param					cursor			query	string	false	"Slack cursor for pagination"
// @Param					limit			query	int		false	"Page size (Slack default 100, max 1000)"
// @Param					types			query	string	false	"Comma-separated channel types (e.g. public_channel,private_channel)"
// @Success				200	{object}	ListChannelsResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/installations/{installation_id}/channels [GET]
func (s *service) ListChannels(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installationID := ctx.Param("installation_id")
	if installationID == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("installation_id is required")))
		return
	}

	limit := defaultChannelsPageLimit
	if v := ctx.Query("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 {
			ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid limit: %q", v)))
			return
		}
		limit = parsed
	}

	install, err := s.getInstallationForOrg(ctx, org.ID, installationID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load slack installation: %w", err))
		return
	}

	resp, err := s.slackClient.ConversationsList(ctx, install.BotAccessToken, slackclient.ConversationsListRequest{
		Cursor:          ctx.Query("cursor"),
		Limit:           limit,
		Types:           ctx.Query("types"),
		ExcludeArchived: true,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("slack conversations.list failed: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ListChannelsResponse{
		Channels:   resp.Channels,
		NextCursor: resp.ResponseMetadata.NextCursor,
	})
}

// getInstallationForOrg loads a SlackInstallation by ID and verifies it is
// reachable from the calling org via a verified SlackOrgLink. ABAC at the DB
// query level — no row returned means either the install doesn't exist or
// the org isn't authorized to see it. We don't distinguish the two so we
// don't leak existence of installations across orgs.
func (s *service) getInstallationForOrg(ctx context.Context, orgID, installationID string) (*app.SlackInstallation, error) {
	linkTable := app.SlackOrgLink{}.TableName()

	var install app.SlackInstallation
	res := s.db.WithContext(ctx).
		Joins(
			"JOIN "+linkTable+" ON "+linkTable+".team_id = slack_installations.team_id"+
				" AND "+linkTable+".org_id = ?"+
				" AND "+linkTable+".status = ?"+
				" AND "+linkTable+".deleted_at = 0",
			orgID, app.SlackOrgLinkStatusVerified,
		).
		Where(app.SlackInstallation{
			ID:     installationID,
			Status: app.SlackInstallationStatusActive,
		}).
		First(&install)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{Err: fmt.Errorf("slack installation %q not found", installationID)}
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &install, nil
}
