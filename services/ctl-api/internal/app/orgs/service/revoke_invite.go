package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						RevokeOrgInvite
// @Summary				Revoke an org invite
// @Description.markdown	revoke_org_invite.md
// @Param					invite_id	path	string	true	"invite ID"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.OrgInvite
// @Router					/v1/orgs/current/invites/{invite_id}/revoke [POST]
func (s *service) RevokeOrgInvite(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.isOrgAdmin(acct, org.ID) {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can revoke invites"),
			Description: "only org admins can revoke invites",
		})
		return
	}

	inviteID := ctx.Param("invite_id")
	if inviteID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invite_id is required"),
			Description: "invite_id is required",
		})
		return
	}

	invite, err := s.getOrgInviteByID(ctx, org.ID, inviteID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get invite: %w", err))
		return
	}

	if invite.Status != app.OrgInviteStatusPending {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("only pending invites can be revoked"),
			Description: "only pending invites can be revoked",
		})
		return
	}

	invite.Status = app.OrgInviteStatusRevoked
	if err := s.db.WithContext(ctx).Save(invite).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to revoke invite: %w", err))
		return
	}

	if err := s.db.WithContext(ctx).Delete(invite).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to soft-delete invite: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, invite)

	s.analyticsClient.Track(ctx, events.InviteRevoked, map[string]interface{}{
		"invite_id": invite.ID,
		"email":     invite.Email,
		"org_id":    invite.OrgID,
		"role_type": invite.RoleType,
	})
}

func (s *service) isOrgAdmin(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}
