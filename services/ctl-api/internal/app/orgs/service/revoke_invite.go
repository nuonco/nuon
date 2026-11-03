package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
func (s *service) RevokeOrgInvite(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}

	inviteID := ctx.Param("invite_id")
	if inviteID == "" {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("invite_id is required"),
			Description: "invite_id is required",
		}
	}

	tx := access.DB()
	var invite app.OrgInvite
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.OrgInvite{ID: inviteID, OrgID: access.Target().OrgID}).Take(&invite).Error; err != nil {
		return require.Response{}, err
	}
	if invite.Status != app.OrgInviteStatusPending {
		return require.Response{}, stderr.ErrUser{Err: fmt.Errorf("only pending invites can be revoked"), Description: "only pending invites can be revoked"}
	}
	if err := tx.Model(&invite).Update("status", app.OrgInviteStatusRevoked).Error; err != nil {
		return require.Response{}, err
	}
	if err := tx.Delete(&invite).Error; err != nil {
		return require.Response{}, err
	}

	ctx.Set("org-invite-mutation", inviteMutation{Invite: invite, Event: events.InviteRevoked})
	return require.Response{Status: http.StatusOK, Body: invite}, nil
}

func (s *service) isOrgAdmin(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}
