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

// @ID						DeleteSlackOrgLink
// @Summary				Revoke a Slack workspace ↔ org binding
// @Description			Soft-deletes the SlackOrgLink. Channel subscriptions cascade off via the FK. Idempotent if the link is already revoked.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Param					link_id	path	string	true	"Slack org link ID"
// @Success				204
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/org-links/{link_id} [DELETE]
func (s *service) DeleteOrgLink(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	linkID := ctx.Param("link_id")
	if linkID == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("link_id is required")))
		return
	}

	if err := s.deleteOrgLink(ctx, org.ID, linkID); err != nil {
		ctx.Error(fmt.Errorf("unable to delete slack org link: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) deleteOrgLink(ctx context.Context, orgID, linkID string) error {
	// ABAC: the link must belong to the caller's org. We resolve+scope the
	// row in a single query so callers can't delete other orgs' links.
	var link app.SlackOrgLink
	res := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:    linkID,
			OrgID: orgID,
		}).
		First(&link)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return stderr.ErrNotFound{Err: fmt.Errorf("slack org link %q not found", linkID)}
	}
	if res.Error != nil {
		return res.Error
	}

	// PG CASCADE on slack_channel_subscriptions.org_link_id only fires on
	// hard deletes, so we mirror it explicitly for the soft-delete path:
	// soft-delete every channel subscription routed via this link, then
	// soft-delete the link itself. Wrapped in a transaction so a partial
	// failure can't leave dangling subs.
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(app.SlackChannelSubscription{OrgLinkID: link.ID}).
			Delete(&app.SlackChannelSubscription{}).Error; err != nil {
			return fmt.Errorf("unable to soft-delete subscriptions for link: %w", err)
		}
		if err := tx.Delete(&link).Error; err != nil {
			return fmt.Errorf("unable to soft-delete org link: %w", err)
		}
		return nil
	})
}
